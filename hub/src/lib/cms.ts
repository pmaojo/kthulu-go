import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';

const DOCS_DIR = path.join(process.cwd(), '..', 'docs');
const REGISTRY_DIR = path.join(process.cwd(), '..', 'registry');

export interface DocContent {
  slug: string[];
  title: string;
  description?: string;
  content: string;
  frontmatter: Record<string, any>;
}

export function getDocBySlug(slug: string[], baseDir: string = DOCS_DIR): DocContent | null {
  if (!Array.isArray(slug)) return null;
  
  const fullPath = path.join(baseDir, ...slug) + '.md';
  const indexPath = path.join(baseDir, ...slug, 'index.md');
  
  let targetPath = '';
  if (fs.existsSync(fullPath)) targetPath = fullPath;
  else if (fs.existsSync(indexPath)) targetPath = indexPath;
  
  if (!targetPath) return null;

  const fileContents = fs.readFileSync(targetPath, 'utf8');
  const { data, content } = matter(fileContents);

  return {
    slug,
    title: data.title || slug[slug.length - 1],
    description: data.description || '',
    content,
    frontmatter: data,
  };
}

export function getAllDocs(subDir: string = '', baseDir: string = DOCS_DIR): DocContent[] {
  const fullDir = path.join(baseDir, subDir);
  if (!fs.existsSync(fullDir)) return [];

  const files = fs.readdirSync(fullDir, { recursive: true }) as string[];
  
  return files
    .filter((file) => file.endsWith('.md'))
    .map((file) => {
      const relativePath = path.join(subDir, file);
      const slug = relativePath.replace(/(\.md|index\.md)$/, '').split(path.sep).filter(Boolean);
      return getDocBySlug(slug, baseDir);
    })
    .filter((doc): doc is DocContent => doc !== null);
}

export function getMarketplaceItems() {
  const starters = getAllDocs('starters', REGISTRY_DIR);
  const modules = getAllDocs('modules', REGISTRY_DIR);
  const plugins = getAllDocs('plugins', REGISTRY_DIR);

  return {
    starters,
    modules,
    plugins,
  };
}
