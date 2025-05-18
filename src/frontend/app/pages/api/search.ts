import type { NextApiRequest, NextApiResponse } from 'next';
import { exec } from 'child_process';
import util from 'util';
import path from 'path';

const execPromise = util.promisify(exec);

interface SearchResult {
  Path: { Ingredients: string[]; Result: string }[];
  VisitedNodes: number;
  ExecutionTime: number;
  TreeStructure: { nodes: any[]; edges: any[]; target?: string; recipes?: any[] };
  VariationIndex?: number;
}

interface ErrorResponse {
  error: string;
}

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse<SearchResult | ErrorResponse>
) {
  if (req.method !== 'GET') {
    return res.status(405).json({ error: 'Method not allowed. Use GET.' });
  }

  const { target, algo, mode, max } = req.query;

  if (!target || !algo || !mode) {
    return res
      .status(400)
      .json({ error: 'Missing required query parameters: target, algo, mode' });
  }

  const maxNum = max ? parseInt(max as string) : 3;
  if (maxNum < 1 || maxNum > 20) {
    return res
      .status(400)
      .json({ error: 'Max must be between 1 and 20' });
  }

  try {
    const goDir = path.join(
      process.cwd(),
      '..', '..', 'backend', 'src', 'Algorithm'
    );

    const goFiles = [
      'main.go',
      'bfs.go',
      'dfs.go',
      'bid.go'
    ].map(file => path.join(goDir, file)).join(' ');

    const command = `go run ${goFiles} -target="${target}" -algo="${algo}" -mode="${mode}" -max=${maxNum}`;

    const { stdout, stderr } = await execPromise(command, { cwd: goDir });

    if (stderr) {
      console.error('Go program stderr:', stderr);
    }

    const data: SearchResult = JSON.parse(stdout);

    if (!data.TreeStructure || !data.TreeStructure.nodes || !data.TreeStructure.edges) {
      return res
        .status(500)
        .json({ error: 'Invalid response format from backend' });
    }

    res.status(200).json(data);
  } catch (error: any) {
    if (error.stderr && error.stderr.includes('undefined: NewBreadthFirstFinder')) {
      return res.status(500).json({
        error: 'Backend compilation error: Missing function definitions. Check bfs.go, dfs.go, and bid.go.'
      });
    }
    if (error.stderr && error.stderr.includes('ErrElementNotFound')) {
      return res.status(400).json({ error: 'Target element not found' });
    }
    if (error.stderr && error.stderr.includes('ErrNoPathFound')) {
      return res.status(404).json({ error: 'No recipe path found for the target element' });
    }
    if (error.message.includes('no such file')) {
      return res.status(500).json({ error: 'Backend files (e.g., main.go or elements.json) are missing' });
    }

    console.error('Error executing Go program:', error);
    res.status(500).json({ error: 'Failed to execute search: ' + error.message });
  }
}