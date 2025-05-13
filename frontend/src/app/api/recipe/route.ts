import { NextRequest, NextResponse } from 'next/server';
import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';
import { promises as fs } from 'fs';

const execAsync = promisify(exec);

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const { target, algorithm, mode, maxPaths = 3 } = body;

    if (!target || !algorithm) {
      return NextResponse.json({ error: 'Missing required parameters' }, { status: 400 });
    }

    // Use a more reliable way to construct the backend path
    const cwd = process.cwd();
    console.log("Current working directory:", cwd);
    
    // Construct the correct backend path - going up from the frontend directory to the project root
    let backendPath = path.join(cwd, '..', '..', 'backend', 'src', 'Algorithm');
    
    // Verify the path exists before executing
    try {
      await fs.access(backendPath);
      console.log(`Backend path exists: ${backendPath}`);
    } catch (err) {
      console.error(`Backend path does not exist: ${backendPath}`);
      
      // Try an alternative path if the first one fails
      const altBackendPath = path.join(cwd, '..', '..', 'backend', 'src', 'Algorithm');
      try {
        await fs.access(altBackendPath);
        console.log(`Alternative backend path exists: ${altBackendPath}`);
        backendPath = altBackendPath;
      } catch (err) {
        return NextResponse.json({ 
          error: 'Backend path not found. Please check the server configuration.' 
        }, { status: 500 });
      }
    }
    
    // Map algorithm names to numeric codes expected by the Go program
    let algoArg = "1"; // Default to BFS
    if (algorithm.toLowerCase() === "dfs") {
      algoArg = "2";
    } else if (algorithm.toLowerCase() === "bidirectional") {
      algoArg = "3";
    }
    
    // Map mode to numeric code
    let modeArg = mode === "single" ? "1" : "2";
    
    // Create a temporary input file to simulate user input for the Go prograam
    const inputFilePath = path.join(backendPath, "temp_input.txt");
    const inputContent = `${target}\n${modeArg}\n${algoArg}\n${mode === "multiple" ? maxPaths + "\n" : ""}`;
    
    await fs.writeFile(inputFilePath, inputContent);
    
    // Prepare command to run the Go program, using < to pipe the input file
    // On Windows we need to use a different approach
    const isWindows = process.platform === "win32";
    
    let command;
    if (isWindows) {
      // Using type command in Windows to pipe file content
      command = `cd "${backendPath}" && go run main.go bfs.go dfs.go bid.go < temp_input.txt`;
    } else {
      // Unix-style input redirection
      command = `cd "${backendPath}" && go run main.go bfs.go dfs.go bid.go < "${inputFilePath}"`;
    }
    
    console.log(`Executing: ${command}`);
    
    // Run the Go program
    const { stdout, stderr } = await execAsync(command);
    
    // Clean up the temporary input file
    try {
      await fs.unlink(inputFilePath);
    } catch (err) {
      console.warn("Could not delete temporary input file:", err);
    }
    
    if (stderr && !stderr.includes("Loaded")) {
      console.error('Error from Go backend:', stderr);
    }
    
    console.log("Go program output:", stdout);
    
    // Extract important information from the output
    const recipePath = [];
    let visitedNodes = 0;
    let executionTime = 0;
    
    // Regex patterns for different parts of the output
    const visitedNodesRegex = /Visited (\d+) nodes during search/;
    const executionTimeRegex = /Algorithm execution time: (\d+) ms/;
    const pathStepRegex = /\d+: (.+?) \(T\d+\) \+ (.+?) \(T\d+\) → (.+?) \(T\d+\)/g;
    
    // Extract visited nodes count
    const visitedMatch = stdout.match(visitedNodesRegex);
    if (visitedMatch) {
      visitedNodes = parseInt(visitedMatch[1]);
    }
    
    // Extract execution time
    const executionMatch = stdout.match(executionTimeRegex);
    if (executionMatch) {
      executionTime = parseInt(executionMatch[1]);
    }
    
    // Extract path steps
    let match;
    while ((match = pathStepRegex.exec(stdout)) !== null) {
      recipePath.push({
        ingredients: [match[1], match[2]],
        result: match[3]
      });
    }
    
    // If the regex didn't find any steps, try a simpler pattern
    if (recipePath.length === 0) {
      const simpleStepRegex = /\d+: (.+?) \+ (.+?) → (.+)/g;
      while ((match = simpleStepRegex.exec(stdout)) !== null) {
        recipePath.push({
          ingredients: [match[1], match[2]],
          result: match[3]
        });
      }
    }
    
    // If we still couldn't parse the path, return the raw output for debugging
    if (recipePath.length === 0 && stdout.trim() !== '') {
      return NextResponse.json({
        error: 'Failed to parse Go output',
        details: stdout.trim(),
        rawOutput: stdout
      }, { status: 500 });
    }

    // Parse tree structure from the recipe tree section
    const treeSection = stdout.substring(stdout.indexOf("Recipe Tree (Target → Basic Elements):"));
    
    // Create a complete tree structure by parsing the tree output
    const treeStructure = parseTreeStructure(treeSection, recipePath);

    return NextResponse.json({
        path: recipePath,
        visitedNodes,
        executionTime,
        treeStructure,
        // Include the detailed path information
        detailedPath: treeStructure.detailedPath || [],
        // Also include the full recipe path with all discovered combinations
        fullRecipePath: treeStructure.fullRecipePath || []
      });
    
  } catch (error) {
    console.error('Error processing recipe request:', error);
    return NextResponse.json({ 
      error: 'Internal server error', 
      details: error instanceof Error ? error.message : String(error)
    }, { status: 500 });
  }
}

// Helper function to parse the tree structure from the CLI output
function parseTreeStructure(treeSection: string, recipePath: any[]) {
    const nodes = [];
    const edges = [];
    const nodeSet = new Set();
    const tierMap = new Map(); // Map to store tier information for each element
    
    // Create a more detailed path that includes intermediate steps
    const detailedPath = [];
    
    // Parse the entire tree structure to capture the complete search path
    const lines = treeSection.split('\n');
    let indentLevel = 0;
    let nodeStack = [];
    let currentParent = null;
    
    // Start by finding the target element (root of tree)
    const rootMatch = /└── (.+?) \(T(\d+)\)/.exec(treeSection);
    if (rootMatch) {
      const [_, nodeName, tierStr] = rootMatch;
      const tier = parseInt(tierStr);
      
      nodes.push({
        id: nodeName,
        label: nodeName,
        type: "target",
        tier: tier,
        imageUrl: ""
      });
      
      nodeSet.add(nodeName);
      tierMap.set(nodeName, tier);
      nodeStack.push({ name: nodeName, tier: tier });
      currentParent = nodeName;
    }
    
    // Track all recipe combinations found in the tree
    const allRecipes = [];
    let currentCombining = null;
    
    // Process the entire tree section
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      
      // Skip empty lines
      if (!line.trim()) continue;
      
      // Count the indent level by counting leading spaces
      const lineIndent = line.search(/\S|$/);
      
      // Check for "Combining:" lines which indicate parent elements
      if (line.includes('├ Combining:')) {
        if (nodeStack.length > 0) {
          currentCombining = nodeStack[nodeStack.length - 1].name;
        }
        continue;
      }
      
      // Line with element node - match both regular and basic elements
      const nodeMatch = /([├└]── )(.+?) \(T(\d+)(?:, BASIC)?\)/.exec(line);
      if (nodeMatch) {
        const [_, prefix, nodeName, tierStr] = nodeMatch;
        const tier = parseInt(tierStr);
        const isBasic = line.includes("BASIC");
        
        // Adjust stack based on indent level
        while (nodeStack.length > 0 && lineIndent <= indentLevel) {
          nodeStack.pop();
          indentLevel -= 2; // Assuming 2 spaces per indent level
          
          if (nodeStack.length > 0) {
            currentParent = nodeStack[nodeStack.length - 1].name;
          } else {
            currentParent = null;
          }
        }
        
        // Only add if not already in the set
        if (!nodeSet.has(nodeName)) {
          nodes.push({
            id: nodeName,
            label: nodeName,
            type: isBasic ? "basic" : "ingredient",
            tier: tier,
            imageUrl: ""
          });
          
          nodeSet.add(nodeName);
          tierMap.set(nodeName, tier);
        }
        
        // Add edge if we have a parent and this is not under a combining section
        if (currentParent && currentParent !== nodeName && currentCombining !== currentParent) {
          // Only add the edge if it doesn't already exist
          if (!edges.some(e => e.source === nodeName && e.target === currentParent)) {
            edges.push({
              id: `${nodeName}-${currentParent}`,
              source: nodeName,
              target: currentParent
            });
          }
        } 
        // Handle combining sections - these represent recipe ingredients
        else if (currentCombining && currentCombining !== nodeName) {
          // Find or create a recipe entry
          let recipe = allRecipes.find(r => r.result === currentCombining);
          if (!recipe) {
            recipe = { result: currentCombining, ingredients: [] };
            allRecipes.push(recipe);
          }
          
          // Add this node as an ingredient if not already there
          if (!recipe.ingredients.includes(nodeName)) {
            recipe.ingredients.push(nodeName);
          }
          
          // Add edge from ingredient to result
          if (!edges.some(e => e.source === nodeName && e.target === currentCombining)) {
            edges.push({
              id: `${nodeName}-${currentCombining}`,
              source: nodeName,
              target: currentCombining
            });
          }
          
          // If we have a complete recipe (2 ingredients), add it to detailed path
          if (recipe.ingredients.length === 2 && !detailedPath.some(p => 
            p.result === recipe.result && 
            p.ingredients.includes(recipe.ingredients[0]) && 
            p.ingredients.includes(recipe.ingredients[1]))) {
            detailedPath.push({
              ingredients: [...recipe.ingredients],
              result: recipe.result
            });
          }
        }
        
        nodeStack.push({ name: nodeName, tier: tier });
        currentParent = nodeName;
        indentLevel = lineIndent;
        continue;
      }
      
      // Line with "cycle detected" - handle appropriately
      const cycleMatch = /([├└]── )(.+?) \(cycle detected\)/.exec(line);
      if (cycleMatch) {
        const [_, prefix, nodeName] = cycleMatch;
        
        // Don't need to add the node again as it's a cycle reference
        // but we should consider it for recipe connections
        if (currentCombining && currentCombining !== nodeName) {
          // Find or create a recipe entry
          let recipe = allRecipes.find(r => r.result === currentCombining);
          if (!recipe) {
            recipe = { result: currentCombining, ingredients: [] };
            allRecipes.push(recipe);
          }
          
          // Add this node as an ingredient if not already there
          if (!recipe.ingredients.includes(nodeName)) {
            recipe.ingredients.push(nodeName);
          }
          
          // Add edge from ingredient to result
          if (!edges.some(e => e.source === nodeName && e.target === currentCombining)) {
            edges.push({
              id: `${nodeName}-${currentCombining}`,
              source: nodeName,
              target: currentCombining
            });
          }
        }
      }
    }
    
    // Organize the detailed path to show the sequence of combinations
    // Start from the basic elements and work up to the target
    const sortedDetailedPath = [];
    const processedResults = new Set();
    
    // First add any recipes that create direct ingredients for the target
    const finalRecipe = recipePath[recipePath.length - 1];
    if (finalRecipe) {
      // Find recipes that create each ingredient of the final recipe
      for (const ingredient of finalRecipe.ingredients) {
        const ingredientRecipes = allRecipes.filter(r => r.result === ingredient);
        
        for (const recipe of ingredientRecipes) {
          if (!processedResults.has(recipe.result)) {
            sortedDetailedPath.push({
              ingredients: recipe.ingredients,
              result: recipe.result
            });
            processedResults.add(recipe.result);
          }
        }
      }
      
      // Add the final recipe if not already added
      if (!processedResults.has(finalRecipe.result)) {
        sortedDetailedPath.push(finalRecipe);
        processedResults.add(finalRecipe.result);
      }
    }
    
    // Add any missing recipes from the original path
    for (const step of recipePath) {
      if (!processedResults.has(step.result)) {
        sortedDetailedPath.push(step);
        processedResults.add(step.result);
      }
    }
    
    // Add any remaining recipes from the tree
    for (const recipe of allRecipes) {
      if (!processedResults.has(recipe.result)) {
        sortedDetailedPath.push({
          ingredients: recipe.ingredients,
          result: recipe.result
        });
        processedResults.add(recipe.result);
      }
    }
    
    // Return both the tree structure and more detailed path information
    return { 
      nodes, 
      edges,
      detailedPath: sortedDetailedPath,
      fullRecipePath: detailedPath 
    };
  }