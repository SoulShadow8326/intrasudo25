import discord
from discord.ext import commands
import aiohttp
import asyncio
import json
import os
from dotenv import load_dotenv
import hypercorn.asyncio
from hypercorn.config import Config
from starlette.applications import Starlette
from starlette.routing import Route
from starlette.requests import Request
from starlette.responses import JSONResponse

load_dotenv()

DISCORD_TOKEN = os.getenv('DISCORD_TOKEN')
SOCKET_PATH = os.getenv('SOCKET_PATH', '/tmp/intrasudo25.sock')
BOT_AUTH_TOKEN = os.getenv('DISCORD_BOT_TOKEN')
BOT_SOCKET_PATH = os.getenv('BOT_SOCKET_PATH', '/tmp/discord_bot.sock')

intents = discord.Intents.default()
intents.message_content = True
bot = commands.Bot(command_prefix='!', intents=intents)

lead_channels = {}
hint_channels = {}
discord_msg_to_db = {}
user_msg_to_discord = {}

async def forward_message(request):
    try:
        body = await request.json()
        user_email = body.get('userEmail')
        username = body.get('username')
        message = body.get('message')
        level = body.get('level')
        
        if not all([user_email, username, message, level is not None]):
            return JSONResponse({'error': 'Missing required fields'}, status_code=400)
        
        if level not in lead_channels:
            return JSONResponse({'error': f'No lead channel found for level {level}'}, status_code=404)
        
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
        
        return JSONResponse({'success': True, 'message': 'Message forwarded to Discord', 'discordMsgId': message_id})
    
    except Exception as e:
        print(f'Error forwarding message: {e}')
        return JSONResponse({'error': 'Internal server error'}, status_code=500)

async def refresh_channels(request):
    try:
        for guild in bot.guilds:
            asyncio.run_coroutine_threadsafe(
                setup_channels(guild),
                bot.loop
            )
        return JSONResponse({'success': True, 'message': 'Channels refreshed'})
    except Exception as e:
        print(f'Error refreshing channels: {e}')
        return JSONResponse({'error': 'Internal server error'}, status_code=500)

app = Starlette(routes=[
    Route('/discord/forward', forward_message, methods=['POST']),
    Route('/discord/refresh', refresh_channels, methods=['POST']),
])

def get_unix_connector():
    return aiohttp.UnixConnector(path=SOCKET_PATH)

@bot.command(name='lock')
async def lock_chat(ctx, level=None):
    if not ctx.author.guild_permissions.administrator:
        await ctx.send("You need administrator permissions to use this command.")
        return
    
    try:
        headers = {
            'Authorization': f'Bearer {BOT_AUTH_TOKEN}',
            'Content-Type': 'application/json'
        }
        
        if level == "all":
            data = {'status': 'locked'}
            endpoint = 'http://unix/api/discord/chat/status'
        elif level and level.isdigit():
            level_num = int(level)
            data = {'status': 'locked', 'level': level_num}
            endpoint = 'http://unix/api/discord/chat/level/status'
        else:
            await ctx.send("Usage: !lock <level_number> or !lock all")
            return
        
        async with aiohttp.ClientSession(connector=get_unix_connector()) as session:
            async with session.post(endpoint, json=data, headers=headers) as response:
                if response.status == 200:
                    target = f"level {level}" if level != "all" else "all levels"
                    await ctx.send(f"Chat locked for {target}")
                    print(f"Discord: Chat locked for {target}")
                else:
                    await ctx.send(f"Failed to lock chat: {response.status}")
                    print(f"Discord: Failed to lock chat for {target}: {response.status}")
    except Exception as e:
        await ctx.send(f"Error: {str(e)}")

@bot.command(name='active')
async def activate_chat(ctx, level=None):
    if not ctx.author.guild_permissions.administrator:
        await ctx.send("You need administrator permissions to use this command.")
        return
    
    try:
        headers = {
            'Authorization': f'Bearer {BOT_AUTH_TOKEN}',
            'Content-Type': 'application/json'
        }
        
        if level == "all":
            data = {'status': 'active'}
            endpoint = 'http://unix/api/discord/chat/status'
        elif level and level.isdigit():
            level_num = int(level)
            data = {'status': 'active', 'level': level_num}
            endpoint = 'http://unix/api/discord/chat/level/status'
        elif level is None:
            data = {'status': 'active'}
            endpoint = 'http://unix/api/discord/chat/status'
        else:
            await ctx.send("Usage: !active <level_number>, !active all, or !active")
            return
        
        async with aiohttp.ClientSession(connector=get_unix_connector()) as session:
            async with session.post(endpoint, json=data, headers=headers) as response:
                if response.status == 200:
                    if level is None:
                        target = "all levels"
                    elif level == "all":
                        target = "all levels"
                    else:
                        target = f"level {level}"
                    await ctx.send(f"Chat activated for {target}")
                    print(f"Discord: Chat activated for {target}")
                else:
                    target = f"level {level}" if level and level != "all" else "all levels"
                    await ctx.send(f"Failed to activate chat: {response.status}")
                    print(f"Discord: Failed to activate chat for {target}: {response.status}")
    except Exception as e:
        await ctx.send(f"Error: {str(e)}")

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
    levels = await get_all_levels()
    levels_set = set(levels)
    
    existing_discord_channels = {}
    channels_to_delete = []
    
    for channel in guild.channels:
        if isinstance(channel, discord.TextChannel):
            if channel.name.startswith('lead-level-'):
                try:
                    level_num = int(channel.name.split('-')[-1])
                    if level_num in levels_set:
                        lead_channels[level_num] = channel
                        existing_discord_channels[f'lead-{level_num}'] = True
                    else:
                        channels_to_delete.append(channel)
                except ValueError:
                    pass
            elif channel.name.startswith('hint-level-'):
                try:
                    level_num = int(channel.name.split('-')[-1])
                    if level_num in levels_set:
                        hint_channels[level_num] = channel
                        existing_discord_channels[f'hint-{level_num}'] = True
                    else:
                        channels_to_delete.append(channel)
                except ValueError:
                    pass
    
    for channel in channels_to_delete:
        try:
            level_num = int(channel.name.split('-')[-1])
            channel_type = 'lead' if channel.name.startswith('lead-') else 'hint'
            await channel.delete()
            print(f'Deleted {channel_type} channel for level {level_num}')
            if channel_type == 'lead' and level_num in lead_channels:
                del lead_channels[level_num]
            elif channel_type == 'hint' and level_num in hint_channels:
                del hint_channels[level_num]
        except Exception as e:
            print(f'Failed to delete channel {channel.name}: {e}')
    
    for level in levels:
        if f'lead-{level}' not in existing_discord_channels:
            try:
                channel = await guild.create_text_channel(f'lead-level-{level}')
                lead_channels[level] = channel
                print(f'Created lead channel for level {level}')
            except Exception as e:
                print(f'Failed to create lead channel for level {level}: {e}')
        
        if f'hint-{level}' not in existing_discord_channels:
            try:
                channel = await guild.create_text_channel(f'hint-level-{level}')
                hint_channels[level] = channel
                print(f'Created hint channel for level {level}')
            except Exception as e:
                print(f'Failed to create hint channel for level {level}: {e}')

async def get_all_levels():
    try:
        async with aiohttp.ClientSession(connector=get_unix_connector()) as session:
            headers = {
                'Authorization': f'Bearer {BOT_AUTH_TOKEN}',
                'Content-Type': 'application/json'
            }
            async with session.get('http://unix/api/levels', headers=headers) as response:
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
    
    async with aiohttp.ClientSession(connector=get_unix_connector()) as session:
        try:
            async with session.post(f'http://unix/api/{endpoint}', 
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

async def refresh_channels(request):
    try:
        for guild in bot.guilds:
            asyncio.run_coroutine_threadsafe(
                setup_channels(guild),
                bot.loop
            )
        return JSONResponse({'success': True, 'message': 'Channels refreshed'})
    except Exception as e:
        print(f'Error refreshing channels: {e}')
        return JSONResponse({'error': 'Internal server error'}, status_code=500)

async def run_asgi_app():
    config = Config()
    config.bind = [f"unix:{BOT_SOCKET_PATH}"]
    config.use_reloader = False
    config.graceful_timeout = 1
    
    await hypercorn.asyncio.serve(app, config, shutdown_trigger=asyncio.Event().wait)

async def run_discord_bot():
    await bot.start(DISCORD_TOKEN)

async def main():
    # Remove old socket file if it exists
    if os.path.exists(BOT_SOCKET_PATH):
        os.unlink(BOT_SOCKET_PATH)
    
    print(f'ASGI server starting on socket {BOT_SOCKET_PATH}')
    
    # Run both ASGI server and Discord bot concurrently
    await asyncio.gather(
        run_asgi_app(),
        run_discord_bot()
    )

if __name__ == '__main__':
    if not DISCORD_TOKEN:
        print("Please set DISCORD_TOKEN in your .env file")
        exit(1)
    
    if not BOT_AUTH_TOKEN:
        print("Please set DISCORD_BOT_TOKEN in your .env file")
        exit(1)
    
    asyncio.run(main())
