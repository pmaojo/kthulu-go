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
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  frontmatter: Record<string, any>;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function readFrontMatter(filePath: string): Promise<Record<string, any>> {
  const fd = await fs.promises.open(filePath, 'r');
  try {
    const buffer = Buffer.alloc(4096);
    const { bytesRead } = await fd.read(buffer, 0, 4096, 0);
    const content = buffer.toString('utf8', 0, bytesRead);

    if (content.startsWith('---')) {
      const end = content.indexOf('\n---', 3);
      if (end !== -1) {
        return matter(content).data;
      }
    }

    // Optimization: If we read less than buffer size, we have the full file.
    if (bytesRead < 4096) {
      return matter(content).data;
    }

    // Fallback: read full file if frontmatter is huge or not found in first 4KB
    const fullContent = await fs.promises.readFile(filePath, 'utf8');
    return matter(fullContent).data;
  } finally {
    await fd.close();
  }
}

// Optimization: Allow passing knownPath to avoid redundant filesystem checks (guessing .md vs /index.md)
export async function getDocBySlug(slug: string[], baseDir: string = DOCS_DIR, fields: string[] = [], knownPath?: string): Promise<DocContent | null> {
  if (!Array.isArray(slug)) return null;
  
  const fullPath = knownPath || (path.join(baseDir, ...slug) + '.md');
  const indexPath = path.join(baseDir, ...slug, 'index.md');
  
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let data: Record<string, any> = {};
  let content = '';
  const needContent = fields.length === 0 || fields.includes('content');

  const tryRead = async (p: string) => {
    if (!needContent) {
      return { data: await readFrontMatter(p), content: '' };
    } else {
      const fileContents = await fs.promises.readFile(p, 'utf8');
      return matter(fileContents);
    }
  };

  try {
    const result = await tryRead(fullPath);
    data = result.data;
    content = result.content;
  } catch (err: unknown) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    if ((err as any).code !== 'ENOENT') throw err;

    // If knownPath was provided, we don't check other paths
    if (knownPath) return null;

    // If fullPath doesn't exist, try indexPath
    try {
      const result = await tryRead(indexPath);
      data = result.data;
      content = result.content;
    } catch (err2: unknown) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      if ((err2 as any).code !== 'ENOENT') throw err2;
      return null;
    }
  }

  const doc: DocContent = {
    slug,
    title: data.title || slug[slug.length - 1],
    description: data.description || '',
    frontmatter: data,
  };

  if (needContent) {
    doc.content = content;
  }

  return doc;
}

export async function getAllDocs(subDir: string = '', baseDir: string = DOCS_DIR, fields: string[] = []): Promise<DocContent[]> {
  const fullDir = path.join(baseDir, subDir);

  let files: string[] = [];
  try {
    files = await fs.promises.readdir(fullDir, { recursive: true }) as string[];
  } catch (err: unknown) {
    // If directory doesn't exist, return empty array
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    if ((err as any).code === 'ENOENT') return [];
    throw err;
  }
  
  const docs = await Promise.all(files
    .filter((file) => file.endsWith('.md'))
    .map(async (file) => {
      const relativePath = path.join(subDir, file);
      const fullPath = path.join(fullDir, file);
      const slug = relativePath.replace(/(\.md|index\.md)$/, '').split(path.sep).filter(Boolean);
      return getDocBySlug(slug, baseDir, fields, fullPath);
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
