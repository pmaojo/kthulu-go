const fs = require('fs');
const path = require('path');
const readline = require('readline');

function toSlug(name) {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

function toTitle(name) {
  return toSlug(name)
    .split('-')
    .map(part => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

function toClass(name) {
  return toSlug(name)
    .split('-')
    .map(part => part.charAt(0).toUpperCase() + part.slice(1))
    .join('') + 'Page';
}

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

rl.question('Module name: ', (answer) => {
  rl.close();
  if (!answer) {
    console.error('Module name is required.');
    process.exit(1);
  }

  const slug = toSlug(answer);
  const title = toTitle(answer);
  const className = toClass(answer);

  const templatesDir = path.join(__dirname, 'templates');
  const specTemplate = fs.readFileSync(path.join(templatesDir, 'spec.tpl'), 'utf8');
  const pageTemplate = fs.readFileSync(path.join(templatesDir, 'page.tpl'), 'utf8');

  const specContent = specTemplate
    .replace(/\{\{SLUG\}\}/g, slug)
    .replace(/\{\{TITLE\}\}/g, title)
    .replace(/\{\{PAGE_CLASS\}\}/g, className);

  const pageContent = pageTemplate
    .replace(/\{\{SLUG\}\}/g, slug)
    .replace(/\{\{TITLE\}\}/g, title)
    .replace(/\{\{PAGE_CLASS\}\}/g, className);

  const testsDir = path.join(__dirname, '..', 'tests');
  const pagesDir = path.join(__dirname, '..', 'pages');

  const specPath = path.join(testsDir, `${slug}.spec.ts`);
  const pagePath = path.join(pagesDir, `${slug}-page.ts`);

  if (fs.existsSync(specPath) || fs.existsSync(pagePath)) {
    console.error('Spec or page object already exists.');
    process.exit(1);
  }

  fs.writeFileSync(specPath, specContent);
  fs.writeFileSync(pagePath, pageContent);

  console.log(`Created ${path.relative(path.join(__dirname, '..'), specPath)}`);
  console.log(`Created ${path.relative(path.join(__dirname, '..'), pagePath)}`);
});
