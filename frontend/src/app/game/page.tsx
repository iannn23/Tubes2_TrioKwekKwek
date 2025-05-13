"use client";

import { useState, useEffect } from "react";
import elementsData from "../../../../backend/src/Scraper/elements.json";
import RecipeTree from "@/components/RecipeTree";
import { Node, Edge } from "reactflow";

interface SearchResult {
  path: Array<{
    ingredients: string[];
    result: string;
  }>;
  visitedNodes: number;
  executionTime: number;
  treeStructure: {
    nodes: Array<{
      id: string;
      label: string;
      type: string;
      tier: number;
      imageUrl: string;
    }>;
    edges: Array<{
      id: string;
      source: string;
      target: string;
    }>;
    target: string;
    recipes: Array<{
      ingredients: string[];
      result: string;
    }>;
  };
}

function convertToReactFlowNodes(rawNodes: any[]): Node[] {
  return rawNodes.map((n) => ({
    id: n.id,
    data: { label: n.label },
    position: { x: 0, y: 0 }, // akan diatur ulang di RecipeTree
    type: "default",
  }));
}

function convertToReactFlowEdges(rawEdges: any[]): Edge[] {
  return rawEdges.map((e) => ({
    id: e.id,
    source: e.source,
    target: e.target,
    type: "default",
  }));
}

export default function GamePage() {
  const [algorithm, setAlgorithm] = useState("BFS");
  const [mode, setMode] = useState<"single" | "multiple">("single");
  const [maxRecipes, setMaxRecipes] = useState(3);
  const [targetElement, setTargetElement] = useState("");
  const [isSearching, setIsSearching] = useState(false);
  const [validElements, setValidElements] = useState<string[]>([]);
  const [isValidElement, setIsValidElement] = useState(false);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);

  // Search results
  const [searchResult, setSearchResult] = useState<SearchResult | null>(null);
  const [nodeVisitSequence, setNodeVisitSequence] = useState<string[]>([]);
  const [currentVisitIndex, setCurrentVisitIndex] = useState(-1);
  const [nodesDiscovered, setNodesDiscovered] = useState<string[]>([]);
  const [edgesDiscovered, setEdgesDiscovered] = useState<
    { id: string; source: string; target: string }[]
  >([]);
  const [searchStats, setSearchStats] = useState({
    totalNodes: 0,
    executionTime: 0,
  });

  // Extract all element names from the JSON
  useEffect(() => {
    // Extract all element names from all tiers
    const allElements: string[] = [];
    elementsData.forEach((tier) => {
      tier.elements.forEach((element) => {
        if (element.name) {
          allElements.push(element.name);
        }
      });
    });
    setValidElements(allElements);
  }, []);

  // Check if input is valid when it changes
  useEffect(() => {
    if (targetElement.trim() === "") {
      setIsValidElement(false);
      setSuggestions([]);
      return;
    }

    // Check if exact match exists
    const exactMatch = validElements.some(
      (elem) => elem.toLowerCase() === targetElement.toLowerCase()
    );
    setIsValidElement(exactMatch);

    // Generate suggestions for partial matches
    if (!exactMatch && targetElement.length > 1) {
      const matches = validElements
        .filter((elem) =>
          elem.toLowerCase().includes(targetElement.toLowerCase())
        )
        .slice(0, 5); // Limit to 5 suggestions
      setSuggestions(matches);
      setShowSuggestions(matches.length > 0);
    } else {
      setSuggestions([]);
      setShowSuggestions(false);
    }
  }, [targetElement, validElements]);

  // Effect for animating the node discovery process
  useEffect(() => {
    if (
      currentVisitIndex >= 0 &&
      currentVisitIndex < nodeVisitSequence.length
    ) {
      const timer = setTimeout(() => {
        const node = nodeVisitSequence[currentVisitIndex];

        // Add node to discovered list
        setNodesDiscovered((prev) => [...prev, node]);

        // Find edges connected to this node in the full tree
        if (searchResult && searchResult.treeStructure) {
          const newEdges = searchResult.treeStructure.edges
            .filter(
              (edge: { target: string; source: string; id: string }) =>
                (edge.target === node &&
                  nodesDiscovered.includes(edge.source)) ||
                (edge.source === node && nodesDiscovered.includes(edge.target))
            )
            .filter(
              (edge: { id: string }) =>
                !edgesDiscovered.some((e) => e.id === edge.id)
            );

          if (newEdges.length > 0) {
            setEdgesDiscovered((prev) => [...prev, ...newEdges]);
          }
        }

        setCurrentVisitIndex((prev) => prev + 1);
      }, 300); // Animation delay in ms

      return () => clearTimeout(timer);
    } else if (
      currentVisitIndex >= nodeVisitSequence.length &&
      nodeVisitSequence.length > 0
    ) {
      // Animation complete
      console.log("Tree visualization complete");
    }
  }, [
    currentVisitIndex,
    nodeVisitSequence,
    nodesDiscovered,
    searchResult,
    edgesDiscovered,
  ]);

  const selectSuggestion = (suggestion: string) => {
    setTargetElement(suggestion);
    setShowSuggestions(false);
  };

  const handleSearch = async () => {
    if (!targetElement.trim()) {
      alert("Nama elemen tidak boleh kosong!");
      return;
    }

    if (!isValidElement) {
      alert(
        "Elemen tidak valid! Gunakan elemen yang ada dalam database Little Alchemy 2."
      );
      return;
    }

    setIsSearching(true);
    setSearchResult(null);
    setNodesDiscovered([]);
    setEdgesDiscovered([]);
    setNodeVisitSequence([]);
    setCurrentVisitIndex(-1);

    try {
      // Make API call to backend through our Next.js API route
      const response = await fetch(`/api/recipe`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          target: targetElement,
          algorithm: algorithm.toLowerCase(),
          mode: mode === "single" ? "single" : "multiple",
          maxPaths: maxRecipes,
        }),
      });

      if (!response.ok) {
        throw new Error(`API request failed with status ${response.status}`);
      }

      const data = await response.json();

      // Enhance the response with images
      // (This assumes you have the full elements data with images)
      if (data.treeStructure && data.treeStructure.nodes) {
        data.treeStructure.nodes = data.treeStructure.nodes.map((node) => {
          // Find the element in your elements data
          const elementData = findElementByName(node.label);
          if (elementData) {
            return {
              ...node,
              imageUrl: elementData.imageUrl || "", // Add image URL if available
            };
          }
          return node;
        });
      }

      // Set search results
      setSearchResult(data);
      setSearchStats({
        totalNodes: data.visitedNodes,
        executionTime: data.executionTime,
      });

      // Prepare visualization sequence just like before
      if (data.treeStructure && data.treeStructure.nodes) {
        const targetNode = data.treeStructure.nodes.find(
          (n) => n.type === "target"
        );

        const visitSequence = targetNode ? [targetNode.id] : [];

        const nodeChildren = {};

        data.treeStructure.edges.forEach((edge) => {
          if (!nodeChildren[edge.target]) {
            nodeChildren[edge.target] = [];
          }
          nodeChildren[edge.target].push(edge.source);
        });

        const traverseTree = (nodeId, sequence, visited = new Set()) => {
          if (visited.has(nodeId)) {
            return;
          }
          visited.add(nodeId);

          const children = nodeChildren[nodeId] || [];

          children.forEach((child) => {
            if (!visited.has(child)) {
              sequence.push(child);
              traverseTree(child, sequence, visited);
            }
          });
        };

        if (targetNode) {
          traverseTree(targetNode.id, visitSequence, new Set());
        }

        setNodeVisitSequence(visitSequence);
        if (visitSequence.length > 0) {
          setCurrentVisitIndex(0);
        }
      }
    } catch (error) {
      console.error("Error searching for recipe:", error);
      alert("Terjadi kesalahan saat mencari recipe. Silakan coba lagi.");
    } finally {
      setIsSearching(false);
    }
  };

  // Helper function to find an element by name in your elements data
  const findElementByName = (name) => {
    for (const tier of elementsData) {
      for (const element of tier.elements) {
        if (element.name === name) {
          return element;
        }
      }
    }
    return null;
  };

  const generateRecipePath = (treeStructure) => {
    if (!treeStructure || !treeStructure.nodes || !treeStructure.edges) {
      console.log("Missing tree structure data");
      return [];
    }

    // Find target node
    const targetNode = treeStructure.nodes.find(
      (node) => node.type === "target"
    );
    if (!targetNode) {
      console.log("Target node not found in tree structure");
      return [];
    }

    // Create a map of child nodes for each node (ingredients that make up each result)
    const childrenMap = {};
    treeStructure.edges.forEach((edge) => {
      if (!childrenMap[edge.target]) {
        childrenMap[edge.target] = [];
      }
      childrenMap[edge.target].push(edge.source);
    });

    // Create a map of node info
    const nodeInfoMap = {};
    treeStructure.nodes.forEach((node) => {
      nodeInfoMap[node.id] = node;
    });

    // Keep track of all recipes we find
    const allRecipes = [];

    // Track processed nodes to avoid duplicates
    const processedCombinations = new Set();

    // Process recipes where the result is already known, breadth-first to ensure we get the complete path
    const processNodeBFS = () => {
      // Start with the target node
      const queue = [targetNode.id];
      const visited = new Set([targetNode.id]);

      while (queue.length > 0) {
        const currentId = queue.shift();
        const ingredients = childrenMap[currentId] || [];

        // Only process valid recipes with 2 ingredients
        if (ingredients.length === 2) {
          const validIngredients = ingredients.every((id) => nodeInfoMap[id]);

          if (validIngredients) {
            // Create a unique key for this recipe combination
            const sortedIngredients = [...ingredients].sort();
            const comboKey = `${currentId}:${sortedIngredients.join("+")}`;

            // Only add if we haven't processed this combination before
            if (!processedCombinations.has(comboKey)) {
              processedCombinations.add(comboKey);

              allRecipes.push({
                result: currentId,
                resultTier: nodeInfoMap[currentId]?.tier || 0,
                ingredients: ingredients.map((id) => ({
                  name: id,
                  tier: nodeInfoMap[id]?.tier || 0,
                })),
              });
            }
          }
        }

        // Add all unvisited ingredients to the queue for BFS traversal
        if (ingredients.length > 0) {
          for (const ingredient of ingredients) {
            if (!visited.has(ingredient) && nodeInfoMap[ingredient]) {
              visited.add(ingredient);
              queue.push(ingredient);
            }
          }
        }
      }
    };

    // Start BFS traversal
    processNodeBFS();

    // Sort steps from highest tier (target) to lowest tier (basic elements)
    return allRecipes.sort((a, b) => {
      // First sort by tier descending
      if (b.resultTier !== a.resultTier) {
        return b.resultTier - a.resultTier;
      }

      // If same tier, sort by result name
      return a.result.localeCompare(b.result);
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-950 via-green-900 to-green-800 text-white p-6">
      <h1 className="text-4xl font-bold mb-6 text-green-400">
        Pencarian Recipe Little Alchemy 2
      </h1>

      {/* Element Input Section */}
      <div className="mb-8">
        <h2 className="text-2xl font-semibold mb-4">Masukkan Nama Elemen</h2>
        <div className="flex flex-col gap-2 relative">
          <input
            type="text"
            value={targetElement}
            onChange={(e) => setTargetElement(e.target.value)}
            placeholder="Contoh: Brick, Metal, Plant..."
            className={`bg-gray-900 border ${
              targetElement && !isValidElement
                ? "border-red-500"
                : "border-green-500"
            } text-white rounded-md px-4 py-2 w-full max-w-md`}
          />

          {/* Show suggestions dropdown */}
          {showSuggestions && (
            <div className="absolute top-full left-0 mt-1 w-full max-w-md bg-gray-800 border border-green-600 rounded-md shadow-lg z-10">
              {suggestions.map((suggestion, index) => (
                <div
                  key={index}
                  className="px-4 py-2 hover:bg-green-900 cursor-pointer"
                  onClick={() => selectSuggestion(suggestion)}
                >
                  {suggestion}
                </div>
              ))}
            </div>
          )}

          {targetElement && !isValidElement && (
            <p className="text-red-400 text-sm">
              Elemen tidak ditemukan. Pilih elemen yang valid dari Little
              Alchemy 2.
            </p>
          )}
        </div>
      </div>

      <h2 className="text-2xl font-semibold mb-4">Pilih Algoritma</h2>

      <div className="flex flex-wrap gap-4 mb-6">
        {["BFS", "DFS", "BIDIRECTIONAL"].map((algo) => (
          <button
            key={algo}
            onClick={() => setAlgorithm(algo)}
            className={`px-4 py-2 rounded-md font-semibold border transition
              ${
                algorithm === algo
                  ? "bg-green-600 text-white border-green-400"
                  : "bg-gray-800 text-gray-300 border-gray-600 hover:bg-gray-700"
              }
            `}
          >
            {algo}
          </button>
        ))}
      </div>

      <p className="text-lg mb-6">
        Algoritma dipilih:{" "}
        <span className="text-green-400 font-semibold">{algorithm}</span>
      </p>
      <h2 className="text-2xl font-semibold mb-4">🎯 Pilih Mode Pencarian</h2>
      <div className="flex gap-4 mb-4">
        <button
          onClick={() => setMode("single")}
          className={`px-4 py-2 rounded-md font-semibold border transition ${
            mode === "single"
              ? "bg-green-600 text-white border-green-400"
              : "bg-gray-800 text-gray-300 border-gray-600 hover:bg-gray-700"
          }`}
        >
          🔹 Single Recipe
        </button>
        <button
          onClick={() => setMode("multiple")}
          className={`px-4 py-2 rounded-md font-semibold border transition ${
            mode === "multiple"
              ? "bg-green-600 text-white border-green-400"
              : "bg-gray-800 text-gray-300 border-gray-600 hover:bg-gray-700"
          }`}
        >
          🔸 Multiple Recipe
        </button>
      </div>

      {/* Input maksimal recipe jika multiple */}
      {mode === "multiple" && (
        <div className="mb-6">
          <label htmlFor="max" className="block mb-2">
            Jumlah maksimal recipe yang ingin dicari:
          </label>
          <input
            type="number"
            min={1}
            max={10}
            value={maxRecipes}
            onChange={(e) => setMaxRecipes(Number(e.target.value))}
            className="bg-gray-900 border border-green-500 text-white rounded-md px-3 py-2"
          />
        </div>
      )}

      {/* Ringkasan dan Tombol Cari */}
      <div className="mb-10">
        <p className="text-lg mb-4">
          Mode:{" "}
          <span className="text-green-400 font-semibold">
            {mode === "single"
              ? "Satu Recipe Bebas"
              : `Multiple Recipe (max ${maxRecipes})`}
          </span>{" "}
          | Algoritma:{" "}
          <span className="text-green-400 font-semibold">{algorithm}</span>
          {targetElement && (
            <>
              {" "}
              | Target:{" "}
              <span
                className={`font-semibold ${
                  isValidElement ? "text-green-400" : "text-red-400"
                }`}
              >
                {targetElement}
              </span>
            </>
          )}
        </p>

        <button
          onClick={handleSearch}
          disabled={isSearching || !targetElement.trim() || !isValidElement}
          className={`px-6 py-3 rounded-md font-bold text-white transition
            ${
              isSearching || !targetElement.trim() || !isValidElement
                ? "bg-gray-700 cursor-not-allowed"
                : "bg-green-600 hover:bg-green-700 border border-green-400"
            }
          `}
        >
          {isSearching ? "Mencari..." : "🔍 Cari Recipe"}
        </button>
      </div>

      {/* Display search stats if result is available */}
      {searchResult && (
        <div className="bg-black/50 rounded-lg p-4 border border-green-500 mb-6">
          <h3 className="text-xl font-semibold text-green-300 mb-2">
            Hasil Pencarian
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="bg-gray-900/80 p-3 rounded-lg">
              <p className="text-gray-300">
                Jumlah Node yang Dikunjungi:
                <span className="text-white font-semibold ml-2">
                  {searchStats.totalNodes}
                </span>
              </p>
            </div>
            <div className="bg-gray-900/80 p-3 rounded-lg">
              <p className="text-gray-300">
                Waktu Eksekusi:
                <span className="text-white font-semibold ml-2">
                  {searchStats.executionTime} ms
                </span>
              </p>
            </div>
          </div>

          {/* Show discovered path */}
          <div className="mt-4 bg-gray-900/80 p-3 rounded-lg">
            <h4 className="text-green-300 mb-2">Path yang ditemukan:</h4>

            {searchResult && (
              <div className="overflow-auto max-h-96 text-sm">
                {searchResult.treeStructure ? (
                  <>
                    {(() => {
                      const generatedPath = generateRecipePath(
                        searchResult.treeStructure
                      );
                      console.log("Generated path:", generatedPath);

                      if (generatedPath.length === 0) {
                        return (
                          <div className="text-yellow-300 mb-2">
                            No recipe path could be generated.
                            <div className="text-xs mt-1">
                              Nodes:{" "}
                              {searchResult.treeStructure.nodes?.length || 0},
                              Edges:{" "}
                              {searchResult.treeStructure.edges?.length || 0}
                            </div>
                          </div>
                        );
                      }

                      // Display the complete path
                      return (
                        <div className="space-y-2">
                          {generatedPath.map((step, index) => (
                            <div
                              key={index}
                              className="p-2 border border-gray-800 rounded bg-gray-900/50"
                            >
                              <div className="font-mono">
                                <span className="text-green-400 font-semibold">
                                  Step {index + 1}:{" "}
                                </span>
                                {step.result} (T{step.resultTier}) →{" "}
                                {step.ingredients
                                  .map((ing) => `${ing.name} (T${ing.tier})`)
                                  .join(" + ")}
                              </div>
                            </div>
                          ))}
                        </div>
                      );
                    })()}

                    <div className="mt-4 border-t border-gray-700 pt-3">
                      <div className="text-yellow-400 font-semibold mb-2">
                        Using raw path data:
                      </div>
                      <div className="space-y-1">
                        {searchResult.path &&
                          searchResult.path.map((step, index) => (
                            <div
                              key={index}
                              className="p-1 border-b border-gray-800"
                            >
                              <span className="text-yellow-400 font-semibold">
                                Raw Step {index + 1}:{" "}
                              </span>
                              {step.result} → {step.ingredients.join(" + ")}
                            </div>
                          ))}
                      </div>
                    </div>
                  </>
                ) : (
                  <div className="text-red-300">
                    Tree structure data is missing. Check the API response.
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      )}

      <div className="bg-gray-900 border border-green-500 p-6 rounded-lg shadow-lg">
        <h2 className="text-2xl font-semibold text-green-300 mb-4">
          Visualisasi Tree Elements Recipe
        </h2>
        {searchResult !== null && searchResult.treeStructure !== null && (
          <RecipeTree
            nodes={convertToReactFlowNodes(searchResult.treeStructure.nodes)}
            edges={convertToReactFlowEdges(searchResult.treeStructure.edges)}
          />
        )}
      </div>
    </div>
  );
}
