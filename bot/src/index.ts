import { Client, GatewayIntentBits, Events, REST, Routes, SlashCommandBuilder } from 'discord.js';
import 'dotenv/config';

// Environment variables
const DISCORD_BOT_TOKEN = process.env.DISCORD_BOT_TOKEN;
const DISCORD_APP_ID = process.env.DISCORD_APP_ID;
const DISCORD_GUILD_ID = process.env.DISCORD_GUILD_ID;

if (!DISCORD_BOT_TOKEN) {
  console.error('Error: DISCORD_BOT_TOKEN is not set');
  process.exit(1);
}

if (!DISCORD_APP_ID) {
  console.error('Error: DISCORD_APP_ID is not set');
  process.exit(1);
}

// Create Discord client
const client = new Client({
  intents: [
    GatewayIntentBits.Guilds,
    GatewayIntentBits.GuildMessages,
  ],
});

// Define slash commands
const commands = [
  new SlashCommandBuilder()
    .setName('ping')
    .setDescription('Replies with Pong!')
    .toJSON(),
];

// Register slash commands
async function registerCommands(): Promise<void> {
  const rest = new REST({ version: '10' }).setToken(DISCORD_BOT_TOKEN!);

  try {
    console.log('Started refreshing application (/) commands.');

    if (DISCORD_GUILD_ID) {
      // Register commands to a specific guild (faster for development)
      await rest.put(
        Routes.applicationGuildCommands(DISCORD_APP_ID!, DISCORD_GUILD_ID),
        { body: commands },
      );
      console.log(`Successfully registered commands to guild: ${DISCORD_GUILD_ID}`);
    } else {
      // Register commands globally
      await rest.put(
        Routes.applicationCommands(DISCORD_APP_ID!),
        { body: commands },
      );
      console.log('Successfully registered global commands.');
    }
  } catch (error) {
    console.error('Error registering commands:', error);
  }
}

// Event: Bot is ready
client.once(Events.ClientReady, (readyClient) => {
  console.log(`✅ Bot is ready! Logged in as ${readyClient.user.tag}`);
  console.log(`📡 Serving ${readyClient.guilds.cache.size} guild(s)`);
});

// Event: Interaction (slash command)
client.on(Events.InteractionCreate, async (interaction) => {
  if (!interaction.isChatInputCommand()) return;

  const { commandName } = interaction;

  if (commandName === 'ping') {
    const latency = Date.now() - interaction.createdTimestamp;
    await interaction.reply(`🏓 Pong! Latency: ${latency}ms`);
  }
});

// Start the bot
async function main(): Promise<void> {
  console.log('🤖 Starting VRC Shift Scheduler Bot...');
  
  // Register commands first
  await registerCommands();
  
  // Login to Discord
  await client.login(DISCORD_BOT_TOKEN);
}

main().catch((error) => {
  console.error('Failed to start bot:', error);
  process.exit(1);
});

