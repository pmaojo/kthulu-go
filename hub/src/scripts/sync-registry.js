/* eslint-disable */
const fs = require('fs');
const path = require('path');

const ROOT = path.join(process.cwd(), '..');
const REGISTRY_DIR = path.join(ROOT, 'registry');
const MODULES_REGISTRY_DIR = path.join(REGISTRY_DIR, 'modules');
const INTERNAL_DIR = path.join(ROOT, 'internal');
const TEMPLATES_DIR = path.join(ROOT, 'cmd', 'kthulu', 'templates');

const DESCRIPTIONS = {
  auth: 'Secure JWT authentication with role-based access control.',
  ai: 'Agentic coding assistant and automated code generation.',
  user: 'Comprehensive user management and profile system.',
  inventory: 'Real-time inventory tracking and warehouse management.',
  organization: 'Multi-tenant organization and team management.',
  invoice: 'Automated invoicing and billing system.',
  product: 'Product catalog and SKU management.',
  notifier: 'Multi-channel notification system (Email, Slack, SMS).',
  calendar: 'Event scheduling and calendar integration.',
  verifactu: 'Spanish fiscal compliance (VeriFACTU) system.',
  realtime: 'Live updates and WebSocket communication layer.',
  health: 'System monitoring and health checks.',
  flags: 'Feature flag management and remote configuration.',
  contact: 'Contact management and CRM integration.',
  oauthsso: 'OAuth 2.0 and Single Sign-On integration.',
  audit: 'Audit logging and compliance tracking.',
  notification: 'User notification and alert system.',
  observability: 'Metrics, tracing, and logging infrastructure.',
  resolver: 'Intelligent module dependency resolution.',
  generator: 'Code generation and scaffolding engine.',
  // Laravel parity modules
  mail: 'Multi-provider email service (SMTP, SES, SendGrid).',
  cache: 'High-performance caching (Memory, Redis, Memcached).',
  storage: 'File storage abstraction (Local, S3, GCS, Azure).',
  scheduler: 'Task scheduling with cron-like syntax.',
  events: 'Event-driven architecture with pub/sub pattern.',
  policy: 'Resource-based authorization and gates.',
  rate: 'Rate limiting (Token Bucket, Sliding Window).',
  seeder: 'Database seeding with faker helpers.',
  session: 'Session management (Cookie, Redis, DB).',
  i18n: 'Internationalization and localization.',
  validate: 'Advanced validation rules and messages.',
};

const ICONS = {
  auth: 'Shield',
  ai: 'Zap',
  user: 'User',
  inventory: 'Package',
  organization: 'Briefcase',
  invoice: 'FileText',
  product: 'Box',
  notifier: 'Bell',
  calendar: 'Calendar',
  verifactu: 'CheckSquare',
  realtime: 'Activity',
  health: 'Heart',
  flags: 'Flag',
  contact: 'Users',
  oauthsso: 'Key',
  audit: 'FileSearch',
  notification: 'MessageSquare',
  observability: 'BarChart',
  resolver: 'GitBranch',
  generator: 'Cpu',
  // Laravel parity modules
  mail: 'Mail',
  cache: 'Database',
  storage: 'HardDrive',
  scheduler: 'Clock',
  events: 'Radio',
  policy: 'ShieldCheck',
  rate: 'Gauge',
  seeder: 'Sprout',
  session: 'Cookie',
  i18n: 'Globe',
  validate: 'CheckCircle',
};

function getAllFiles(dirPath, arrayOfFiles = []) {
  if (!fs.existsSync(dirPath)) return arrayOfFiles;
  const files = fs.readdirSync(dirPath);

  files.forEach(function(file) {
    const fullPath = path.join(dirPath, file);
    if (fs.statSync(fullPath).isDirectory()) {
      if (file !== 'vendor' && file !== 'node_modules' && !file.startsWith('.')) {
        arrayOfFiles = getAllFiles(fullPath, arrayOfFiles);
      }
    } else {
      if (file.endsWith('.go') || file.endsWith('.tmpl') || file.endsWith('.md')) {
        arrayOfFiles.push(fullPath);
      }
    }
  });

  return arrayOfFiles;
}

function syncModules() {
  console.log('🔄 Syncing Modules via @kthulu tags...');
  
  const scanDirs = [INTERNAL_DIR, TEMPLATES_DIR];
  const allFiles = [];
  scanDirs.forEach(dir => {
    if (fs.existsSync(dir)) {
      getAllFiles(dir, allFiles);
    }
  });

  const discoveredModules = new Set();

  allFiles.forEach(file => {
    const content = fs.readFileSync(file, 'utf8');
    const matches = content.match(/@kthulu:module:(\w+)/g);
    if (matches) {
      matches.forEach(match => {
        const id = match.split(':').pop();
        if (id) discoveredModules.add(id);
      });
    }
  });

  // Adding manual core modules that are known to exist as templates but might not be tagged yet
  const coreModules = [
    'auth', 'user', 'ai', 'inventory', 'invoice', 'product', 
    'organization', 'notifier', 'calendar', 'realtime', 
    'contact', 'verifactu', 'oauthsso', 'audit', 'notification',
    'health', 'flags', 'observability', 'resolver', 'generator',
    // Laravel parity modules
    'mail', 'cache', 'storage', 'scheduler', 'events', 'policy',
    'rate', 'seeder', 'session', 'i18n', 'validate'
  ];
  coreModules.forEach(id => discoveredModules.add(id));

  console.log(`📦 Registered ${discoveredModules.size} modules: ${Array.from(discoveredModules).sort().join(', ')}`);

  // Create/Update modules
  discoveredModules.forEach(id => {
    const itemDir = path.join(MODULES_REGISTRY_DIR, id);
    if (!fs.existsSync(itemDir)) fs.mkdirSync(itemDir, { recursive: true });

    const metadata = {
      id,
      name: id.charAt(0).toUpperCase() + id.slice(1) + ' Module',
      type: 'module',
      description: DESCRIPTIONS[id] || `High-performance ${id} module for Kthulu Go ecosystems.`,
      author: 'Kthulu Core',
      stars: (id === 'auth' || id === 'ai') ? 100 : 0,
      icon: ICONS[id] || 'Box'
    };

    fs.writeFileSync(path.join(itemDir, 'metadata.json'), JSON.stringify(metadata, null, 2));

    const indexPath = path.join(itemDir, 'index.md');
    const defaultContent = `# ${metadata.name}\n\nThis module was automatically discovered or registered and is ready for use in Kthulu Go.`;
    
    if (!fs.existsSync(indexPath)) {
       fs.writeFileSync(indexPath, `---\ntitle: ${metadata.name}\ndescription: ${metadata.description}\n---\n\n${defaultContent}`);
    } else {
       const existing = fs.readFileSync(indexPath, 'utf8');
       // If it's a minimal placeholder, update it with description
       if (existing.length < 150 || existing.includes('Automatically discovered')) {
          fs.writeFileSync(indexPath, `---\ntitle: ${metadata.name}\ndescription: ${metadata.description}\n---\n\n# ${metadata.name}\n\n${DESCRIPTIONS[id] || 'Detailed documentation coming soon.'}`);
       }
    }
  });

  // Cleanup: Remove ghost modules
  if (fs.existsSync(MODULES_REGISTRY_DIR)) {
    const existingFolders = fs.readdirSync(MODULES_REGISTRY_DIR);
    existingFolders.forEach(folder => {
      if (!discoveredModules.has(folder)) {
        console.log(`🗑️ Removing ghost module: ${folder}`);
        fs.rmSync(path.join(MODULES_REGISTRY_DIR, folder), { recursive: true, force: true });
      }
    });
  }
}

function syncStarters() {
  console.log('🔄 Syncing Starters...');
  const itemDir = path.join(REGISTRY_DIR, 'starters', 'base-api');
  if (!fs.existsSync(itemDir)) fs.mkdirSync(itemDir, { recursive: true });
  
  const metadata = {
     id: 'base-api',
     name: 'Base API Starter',
     type: 'starter',
     description: 'The standard Kthulu Go API project structure.',
     author: 'Kthulu Team',
     stars: 120,
     icon: 'Zap'
  };
  fs.writeFileSync(path.join(itemDir, 'metadata.json'), JSON.stringify(metadata, null, 2));
}

function main() {
  syncModules();
  syncStarters();
  console.log('✅ Registry Sync Complete!');
}

main();
