import fs from 'fs';
import path from 'path';

const ROOT = path.join(process.cwd(), '..');
const REGISTRY_DIR = path.join(ROOT, 'registry');
const BACKEND_MODULES_DIR = path.join(ROOT, 'backend', 'internal', 'adapters', 'http', 'modules');
const CLI_TEMPLATES_DIR = path.join(ROOT, 'backend', 'cmd', 'kthulu-cli', 'templates');

interface RegistryItem {
  id: string;
  name: string;
  type: 'module' | 'starter' | 'plugin';
  description: string;
  author: string;
  stars: number;
  icon: string;
}

function syncModules() {
  console.log('🔄 Syncing Modules from Backend...');
  if (!fs.existsSync(BACKEND_MODULES_DIR)) return;

  const files = fs.readdirSync(BACKEND_MODULES_DIR);
  files.forEach(file => {
    if (file.endsWith('.go') && !file.endsWith('_test.go')) {
      const id = file.replace('.go', '');
      if (id === 'module' || id === 'routes' || id === 'registry') return;

      const itemDir = path.join(REGISTRY_DIR, 'modules', id);
      fs.mkdirSync(itemDir, { recursive: true });

      const metadata: RegistryItem = {
        id,
        name: id.charAt(0).toUpperCase() + id.slice(1) + ' Module',
        type: 'module',
        description: `Automatically discovered Kthulu module: ${id}`,
        author: 'Kthulu Core',
        stars: 0,
        icon: 'Shield'
      };

      fs.writeFileSync(path.join(itemDir, 'metadata.json'), JSON.stringify(metadata, null, 2));
      
      if (!fs.existsSync(path.join(itemDir, 'index.md'))) {
        fs.writeFileSync(path.join(itemDir, 'index.md'), `---\ntitle: ${metadata.name}\ndescription: ${metadata.description}\n---\n\n# ${metadata.name}\n\nThis module was automatically discovered from the Kthulu Monolith.`);
      }
    }
  });
}

function syncStarters() {
  console.log('🔄 Syncing Starters from CLI Templates...');
  // For now, let's look at the scaffold/backend templates as potential starters
  const scaffoldDir = path.join(CLI_TEMPLATES_DIR, 'scaffold', 'backend');
  if (fs.existsSync(scaffoldDir)) {
     // Generic starter
     const itemDir = path.join(REGISTRY_DIR, 'starters', 'base-api');
     fs.mkdirSync(itemDir, { recursive: true });
     const metadata = {
        id: 'base-api',
        name: 'Base API Starter',
        type: 'starter',
        description: 'The standard Kthulu API project structure.',
        author: 'Kthulu Team',
        stars: 10,
        icon: 'Zap'
     };
     fs.writeFileSync(path.join(itemDir, 'metadata.json'), JSON.stringify(metadata, null, 2));
  }
}

function main() {
  syncModules();
  syncStarters();
  console.log('✅ Registry Sync Complete!');
}

main();
