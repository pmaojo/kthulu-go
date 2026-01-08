import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';

const DOCS_DIR = path.join(process.cwd(), '..', 'docs');

export interface DocContent {
  slug: string[];
  title: string;
  description?: string;
  content: string;
  frontmatter: Record<string, any>;
}

export function getDocBySlug(slug: string[]): DocContent | null {
  const fullPath = path.join(DOCS_DIR, ...slug) + '.md';
  
  if (!fs.existsSync(fullPath)) {
    // Try index.md if it's a directory
    const indexPath = path.join(DOCS_DIR, ...slug, 'index.md');
    if (fs.existsSync(indexPath)) {
      const fileContents = fs.readFileSync(indexPath, 'utf8');
      const { data, content } = matter(fileContents);
      return {
        slug,
        title: data.title || slug[slug.length - 1],
        description: data.description || '',
        content,
        frontmatter: data,
      };
    }
    return null;
  }

  const fileContents = fs.readFileSync(fullPath, 'utf8');
  const { data, content } = matter(fileContents);

  return {
    slug,
    title: data.title || slug[slug.length - 1],
    description: data.description || '',
    content,
    frontmatter: data,
  };
}

export function getAllDocs(subDir: string = ''): DocContent[] {
  const fullDir = path.join(DOCS_DIR, subDir);
  if (!fs.existsSync(fullDir)) return [];

  const files = fs.readdirSync(fullDir, { recursive: true }) as string[];
  
  return files
    .filter((file) => file.endsWith('.md'))
    .map((file) => {
      const relativePath = path.join(subDir, file);
      const slug = relativePath.replace(/\.md$/, '').split(path.sep);
      return getDocBySlug(slug);
    })
    .filter((doc): doc is DocContent => doc !== null);
}

export function getMarketplaceItems() {
  const items = getAllDocs('marketplace');
  return {
    starters: items.filter(item => item.frontmatter.type === 'starter'),
    modules: items.filter(item => item.frontmatter.type === 'module'),
    plugins: items.filter(item => item.frontmatter.type === 'plugin'),
  };
}
