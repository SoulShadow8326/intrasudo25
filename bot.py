import discord
from discord.ext import commands
import aiohttp
import asyncio
import json
import os
from dotenv import load_dotenv
from flask import Flask, request, jsonify
import threading

load_dotenv()

DISCORD_TOKEN = os.getenv('DISCORD_TOKEN')
API_BASE_URL = os.getenv('API_BASE_URL', 'http://localhost:8080')
BOT_AUTH_TOKEN = os.getenv('DISCORD_BOT_TOKEN')
BOT_PORT = int(os.getenv('DISCORD_BOT_PORT', '8081'))

intents = discord.Intents.default()
intents.message_content = True
bot = commands.Bot(command_prefix='!', intents=intents)

lead_channels = {}
hint_channels = {}
discord_msg_to_db = {}
user_msg_to_discord = {}

app = Flask(__name__)

@bot.event
async def on_ready():
    print(f'{bot.user} has connected to Discord!')
    
    for guild in bot.guilds:
        print(f'Connected to guild: {guild.name}')
        await setup_channels(guild)

@bot.event
async def on_message_delete(message):
    if message.author == bot.user:
        return
    
    channel = message.channel
    
    if isinstance(channel, discord.TextChannel):
        if channel.name.startswith('hint-level-'):
            await handle_hint_message_deleted(message)

async def setup_channels(guild):
    existing_channels = {}
    
    for channel in guild.channels:
        if isinstance(channel, discord.TextChannel):
            if channel.name.startswith('lead-level-'):
                try:
                    level_num = int(channel.name.split('-')[-1])
                    lead_channels[level_num] = channel
                    existing_channels[f'lead-{level_num}'] = True
                    print(f'Found lead channel for level {level_num}: {channel.name}')
                except ValueError:
                    pass
            elif channel.name.startswith('hint-level-'):
                try:
                    level_num = int(channel.name.split('-')[-1])
                    hint_channels[level_num] = channel
                    existing_channels[f'hint-{level_num}'] = True
                    print(f'Found hint channel for level {level_num}: {channel.name}')
                except ValueError:
                    pass
    
    levels = await get_all_levels()
    
    for level in levels:
        if f'lead-{level}' not in existing_channels:
            try:
                channel = await guild.create_text_channel(f'lead-level-{level}')
                lead_channels[level] = channel
                print(f'Created lead channel for level {level}')
            except Exception as e:
                print(f'Failed to create lead channel for level {level}: {e}')
        
        if f'hint-{level}' not in existing_channels:
            try:
                channel = await guild.create_text_channel(f'hint-level-{level}')
                hint_channels[level] = channel
                print(f'Created hint channel for level {level}')
            except Exception as e:
                print(f'Failed to create hint channel for level {level}: {e}')

async def get_all_levels():
    try:
        async with aiohttp.ClientSession() as session:
            headers = {
                'Authorization': f'Bearer {BOT_AUTH_TOKEN}',
                'Content-Type': 'application/json'
            }
            async with session.get(f'{API_BASE_URL}/api/levels', headers=headers) as response:
                if response.status == 200:
                    data = await response.json()
                    return data.get('levels', [])
                else:
                    print(f'Failed to fetch levels: {response.status}')
                    return []
    except Exception as e:
        print(f'Error fetching levels: {e}')
        return []

async def send_to_api(endpoint, data):
    headers = {
        'Authorization': f'Bearer {BOT_AUTH_TOKEN}',
        'Content-Type': 'application/json'
    }
    
    async with aiohttp.ClientSession() as session:
        try:
            async with session.post(f'{API_BASE_URL}/api/{endpoint}', 
                                  json=data, headers=headers) as response:
                if response.status == 200:
                    result = await response.json()
                    print(f"API response for {endpoint}: {result}")
                    return result
                else:
                    error_text = await response.text()
                    print(f'API Error: {response.status} - {error_text}')
                    return None
        except Exception as e:
            print(f'Request failed: {e}')
            return None

@bot.event
async def on_message(message):
    if message.author == bot.user:
        return
    
    await bot.process_commands(message)
    
    channel = message.channel
    
    if isinstance(channel, discord.TextChannel):
        if channel.name.startswith('lead-level-'):
            if message.reference and message.reference.message_id:
                await handle_lead_message(message)
        elif channel.name.startswith('hint-level-'):
            if not message.content.startswith('**From:**'):
                await handle_hint_message(message)

async def handle_lead_message(message):
    channel_name = message.channel.name
    try:
        level_num = int(channel_name.split('-')[-1])
    except ValueError:
        return
    
    try:
        replied_to_msg = await message.channel.fetch_message(message.reference.message_id)
        if replied_to_msg and replied_to_msg.content.startswith('**From:**'):
            content_lines = replied_to_msg.content.split('\n')
            user_info = content_lines[0].replace('**From:**', '').strip()
            user_email = user_info.split('(')[1].split(')')[0] if '(' in user_info else user_info
            
            data = {
                'type': 'lead_reply',
                'userEmail': user_email,
                'sentBy': message.author.display_name,
                'message': message.content,
                'levelNumber': level_num,
                'discordMsgId': str(message.id),
                'parentMsgId': str(replied_to_msg.id)
            }
            
            response = await send_to_api('discord-bot', data)
            if response and response.get('success'):
                print(f"Successfully forwarded lead reply for level {level_num}")
            else:
                print(f"Failed to forward lead reply for level {level_num}: {response}")
    except Exception as e:
        print(f"Error handling lead reply for level {level_num}: {e}")

async def handle_hint_message(message):
    channel_name = message.channel.name
    try:
        level_num = int(channel_name.split('-')[-1])
    except ValueError:
        return
    
    try:
        data = {
            'type': 'hint_message',
            'message': message.content,
            'sentBy': message.author.display_name,
            'levelNumber': level_num,
            'discordMsgId': str(message.id)
        }
        
        response = await send_to_api('discord-bot', data)
        if response and response.get('success'):
            print(f"Successfully forwarded hint message for level {level_num}")
        else:
            print(f"Failed to forward hint message for level {level_num}: {response}")
    except Exception as e:
        print(f"Error handling hint message for level {level_num}: {e}")

async def handle_hint_message_deleted(message):
    channel_name = message.channel.name
    try:
        level_num = int(channel_name.split('-')[-1])
    except ValueError:
        return
    
    try:
        data = {
            'type': 'hint_message_deleted',
            'discordMsgId': str(message.id)
        }
        
        response = await send_to_api('discord-bot', data)
        if response and response.get('success'):
            print(f"Successfully deleted hint message for level {level_num}")
        else:
            print(f"Failed to delete hint message for level {level_num}: {response}")
    except Exception as e:
        print(f"Error handling hint message deletion for level {level_num}: {e}")

@app.route('/discord/forward', methods=['POST'])
def forward_message():
    try:
        data = request.get_json()
        user_email = data.get('userEmail')
        username = data.get('username')
        message = data.get('message')
        level = data.get('level')
        
        if not all([user_email, username, message, level is not None]):
            return jsonify({'success': False, 'message': 'Missing required fields'}), 400
        
        if level not in lead_channels:
            return jsonify({'success': False, 'message': f'No lead channel found for level {level}'}), 404
        
        channel = lead_channels[level]
        formatted_message = f"**From:** {username} ({user_email})\n{message}"
        future = asyncio.run_coroutine_threadsafe(
            channel.send(formatted_message),
            bot.loop
        )
        
        discord_message = future.result()
        message_id = str(discord_message.id)
        
        user_msg_to_discord[f"{user_email}_{level}_{message[:20]}"] = message_id
        print(f"Created mapping for user message: {user_email}_{level}_{message[:20]} -> Discord ID: {message_id}")
        async def send_update_to_api():
            update_data = {
                'type': 'update_discord_msg_id',
                'userEmail': user_email,
                'message': message[:50],
                'levelNumber': level,
                'discordMsgId': message_id
            }
            result = await send_to_api('discord-bot', update_data)
            if result and result.get('success'):
                db_msg_id = result.get('data', {}).get('id')
                if db_msg_id:
                    discord_msg_to_db[discord_message.id] = db_msg_id
                    print(f"Stored mapping: Discord msg {discord_message.id} -> DB msg {db_msg_id}")
            else:
                print(f"Failed to update message ID in backend: {result}")
        asyncio.run_coroutine_threadsafe(send_update_to_api(), bot.loop)
        
        return jsonify({'success': True, 'message': 'Message forwarded to Discord', 'discordMsgId': message_id})
    
    except Exception as e:
        print(f'Error forwarding message: {e}')
        return jsonify({'success': False, 'message': 'Internal server error'}), 500

@app.route('/discord/refresh', methods=['POST'])
def refresh_channels():
    try:
        for guild in bot.guilds:
            asyncio.run_coroutine_threadsafe(
                setup_channels(guild),
                bot.loop
            )
        return jsonify({'success': True, 'message': 'Channels refreshed'})
    except Exception as e:
        print(f'Error refreshing channels: {e}')
        return jsonify({'success': False, 'message': 'Internal server error'}), 500

async def periodic_channel_refresh():
    while True:
        await asyncio.sleep(300)
        for guild in bot.guilds:
            await setup_channels(guild)

def run_flask_app():
    app.run(host='0.0.0.0', port=BOT_PORT, debug=False)

def run_discord_bot():
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    
    async def start_bot():
        await bot.start(DISCORD_TOKEN)
    
    async def start_tasks():
        await asyncio.gather(
            start_bot(),
            periodic_channel_refresh()
        )
    
    loop.run_until_complete(start_tasks())

if __name__ == '__main__':
    if not DISCORD_TOKEN:
        print("Please set DISCORD_TOKEN in your .env file")
        exit(1)
    
    if not BOT_AUTH_TOKEN:
        print("Please set DISCORD_BOT_TOKEN in your .env file")
        exit(1)
    
    flask_thread = threading.Thread(target=run_flask_app)
    flask_thread.daemon = True
    flask_thread.start()
    
    print(f'Flask server starting on port {BOT_PORT}')
    
    run_discord_bot()
