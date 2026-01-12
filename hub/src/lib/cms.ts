import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';

export const DOCS_DIR = path.join(process.cwd(), '..', 'docs');
export const REGISTRY_DIR = path.join(process.cwd(), '..', 'registry');

export interface DocContent {
  slug: string[];
  title: string;
  description?: string;
  content?: string;
  frontmatter: Record<string, any>;
}

export async function getDocBySlug(slug: string[], baseDir: string = DOCS_DIR, fields: string[] = []): Promise<DocContent | null> {
  if (!Array.isArray(slug)) return null;
  
  const fullPath = path.join(baseDir, ...slug) + '.md';
  const indexPath = path.join(baseDir, ...slug, 'index.md');
  
  let targetPath = '';
  if (fs.existsSync(fullPath)) targetPath = fullPath;
  else if (fs.existsSync(indexPath)) targetPath = indexPath;
  
  if (!targetPath) return null;

  const fileContents = await fs.promises.readFile(targetPath, 'utf8');
  const { data, content } = matter(fileContents);

  const doc: DocContent = {
    slug,
    title: data.title || slug[slug.length - 1],
    description: data.description || '',
    frontmatter: data,
  };

  if (fields.length === 0 || fields.includes('content')) {
    doc.content = content;
  }

  return doc;
}

export async function getAllDocs(subDir: string = '', baseDir: string = DOCS_DIR, fields: string[] = []): Promise<DocContent[]> {
  const fullDir = path.join(baseDir, subDir);
  if (!fs.existsSync(fullDir)) return [];

  const files = await fs.promises.readdir(fullDir, { recursive: true }) as string[];
  
  const docs = await Promise.all(files
    .filter((file) => file.endsWith('.md'))
    .map(async (file) => {
      const relativePath = path.join(subDir, file);
      const slug = relativePath.replace(/(\.md|index\.md)$/, '').split(path.sep).filter(Boolean);
      return getDocBySlug(slug, baseDir, fields);
    }));

  return docs.filter((doc): doc is DocContent => doc !== null);
}

export async function getMarketplaceItems() {
  const fields = ['slug', 'title', 'description', 'frontmatter'];

  const [starters, modules, plugins] = await Promise.all([
    getAllDocs('starters', REGISTRY_DIR, fields),
    getAllDocs('modules', REGISTRY_DIR, fields),
    getAllDocs('plugins', REGISTRY_DIR, fields)
  ]);

  return {
    starters,
    modules,
    plugins,
  };
}
