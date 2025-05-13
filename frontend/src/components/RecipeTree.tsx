"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  Node,
  Edge,
  MarkerType,
  Position,
} from "reactflow";
import "reactflow/dist/style.css";

// Define the props interface for RecipeTree
export interface RecipeTreeProps {
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
  animationInProgress: boolean;
}

// Node type components
const nodeTypes = {
  target: ({ data }: { data: any }) => (
    <div className="p-3 rounded-lg border-2 border-green-400 bg-green-900/50 text-center min-w-[140px]">
      <div className="font-bold text-green-200">{data.label}</div>
      <div className="text-xs text-green-300">Target (T{data.tier})</div>
    </div>
  ),
  basic: ({ data }: { data: any }) => (
    <div className="p-3 rounded-lg border-2 border-blue-400 bg-blue-900/50 text-center min-w-[140px]">
      <div className="font-bold text-blue-200">{data.label}</div>
      <div className="text-xs text-blue-300">Basic (T{data.tier})</div>
    </div>
  ),
  ingredient: ({ data }: { data: any }) => (
    <div className="p-3 rounded-lg border border-gray-400 bg-gray-800/80 text-center min-w-[140px]">
      <div className="font-semibold text-white">{data.label}</div>
      <div className="text-xs text-gray-300">T{data.tier}</div>
    </div>
  ),
  combining: ({ data }: { data: any }) => (
    <div className="p-2 rounded-lg border border-blue-400/40 bg-blue-950/30 text-center min-w-[100px]">
      <div className="text-xs text-blue-300">Combining</div>
    </div>
  ),
};

export default function RecipeTree({ nodes = [], edges = [], animationInProgress = false }: RecipeTreeProps) {
  const [isProcessing, setIsProcessing] = useState(false);
  
  // Optimized calculation of node positions
  const calculateNodePositions = useCallback(() => {
    if (!nodes.length) return [];
    setIsProcessing(true);
    
    try {
      // Find target node as root
      const targetNode = nodes.find(node => node.type === "target");
      if (!targetNode) {
        setIsProcessing(false);
        return nodes.map(node => ({
          id: node.id,
          type: node.type === "basic" ? "basic" : 
                node.type === "target" ? "target" : "ingredient",
          data: { label: node.label, tier: node.tier },
          position: { x: 0, y: 0 },
          sourcePosition: Position.Bottom,
          targetPosition: Position.Top
        }));
      }
      
      // Create node lookup for faster access
      const nodeMap = new Map();
      nodes.forEach(node => {
        nodeMap.set(node.id, node);
      });
      
      // Build graph structure
      const childrenMap = {};
      const parentMap = {};
      
      edges.forEach(edge => {
        // Target <- Source relationship (in recipe tree)
        if (!childrenMap[edge.target]) {
          childrenMap[edge.target] = [];
        }
        childrenMap[edge.target].push(edge.source);
        
        // Track parent relationships
        parentMap[edge.source] = edge.target;
      });
      
      // Track visited nodes to handle cycles
      const visited = new Set();
      const processedNodes = new Set();
      const finalNodes = [];
      
      // Get tier information - nodes at same tier should be on same level
      const tierToNodesMap = new Map();
      nodes.forEach(node => {
        if (!tierToNodesMap.has(node.tier)) {
          tierToNodesMap.set(node.tier, []);
        }
        tierToNodesMap.get(node.tier).push(node.id);
      });
      
      // Sort tiers in descending order (highest tier at top)
      const tiers = [...tierToNodesMap.keys()].sort((a, b) => b - a);
      
      // Layout constants
      const LEVEL_HEIGHT = 150;
      const NODE_WIDTH = 160;
      const NODE_MARGIN = 40;
      
      // Breadth-first processing to assign positions level by level
      const processedEdges = new Set();
      
      // Process each tier level by level
      let yPos = 0;
      
      tiers.forEach(tier => {
        const nodesInTier = tierToNodesMap.get(tier) || [];
        let xPos = -(nodesInTier.length * (NODE_WIDTH + NODE_MARGIN)) / 2;
        
        // Position nodes in current tier
        nodesInTier.forEach(nodeId => {
          const node = nodeMap.get(nodeId);
          if (!node) return;
          
          finalNodes.push({
            id: nodeId,
            type: node.type === "basic" ? "basic" : 
                  node.type === "target" ? "target" : "ingredient",
            data: { 
              label: node.label,
              tier: node.tier
            },
            position: { 
              x: xPos, 
              y: yPos 
            },
            sourcePosition: Position.Bottom,
            targetPosition: Position.Top
          });
          
          processedNodes.add(nodeId);
          xPos += NODE_WIDTH + NODE_MARGIN;
        });
        
        // Increment vertical position for next tier
        yPos += LEVEL_HEIGHT;
      });
      
      // Add combining nodes between parent-children
      const combiningNodes = [];
      const combiningEdges = [];
      
      edges.forEach(edge => {
        const sourceNode = nodeMap.get(edge.source);
        const targetNode = nodeMap.get(edge.target);
        
        if (sourceNode && targetNode && !processedEdges.has(edge.id)) {
          processedEdges.add(edge.id);
          
          // Check if a combining node is needed
          const siblings = childrenMap[edge.target] || [];
          if (siblings.length >= 2) {
            // Create combining node between parent and children
            const combiningId = `combining_${edge.target}`;
            
            // Find parent node in finalNodes for position
            const parentNode = finalNodes.find(n => n.id === edge.target);
            if (parentNode) {
              const combiningNodeExists = combiningNodes.some(n => n.id === combiningId);
              
              if (!combiningNodeExists) {
                // Place combining node halfway between parent and children
                combiningNodes.push({
                  id: combiningId,
                  type: "combining",
                  data: { label: "Combining" },
                  position: { 
                    x: parentNode.position.x,
                    y: parentNode.position.y + (LEVEL_HEIGHT / 2)
                  },
                  sourcePosition: Position.Bottom,
                  targetPosition: Position.Top
                });
                
                // Add edge from parent to combining node
                combiningEdges.push({
                  id: `edge_${edge.target}_to_${combiningId}`,
                  source: edge.target,
                  target: combiningId,
                  type: 'smoothstep'
                });
              }
              
              // Add edge from combining node to child
              combiningEdges.push({
                id: `edge_${combiningId}_to_${edge.source}`,
                source: combiningId,
                target: edge.source,
                type: 'smoothstep'
              });
            }
          } else {
            // Direct connection for single child
            combiningEdges.push({
              id: edge.id,
              source: edge.source,
              target: edge.target,
              type: 'smoothstep'
            });
          }
        }
      });
      
      // Combine regular nodes with combining nodes
      const allNodes = [...finalNodes, ...combiningNodes];
      
      setIsProcessing(false);
      return allNodes;
      
    } catch (error) {
      console.error("Error calculating node positions:", error);
      setIsProcessing(false);
      
      // Fallback to simple layout
      return nodes.map(node => ({
        id: node.id,
        type: node.type === "basic" ? "basic" : 
              node.type === "target" ? "target" : "ingredient",
        data: { label: node.label, tier: node.tier },
        position: { x: 0, y: node.tier * 100 }, // Simple vertical layout by tier
        sourcePosition: Position.Bottom,
        targetPosition: Position.Top
      }));
    }
  }, [nodes, edges]);

  // Use memoized values to reduce computation
  const initialNodes = useMemo(() => {
    if (nodes.length === 0) return [];
    return calculateNodePositions();
  }, [calculateNodePositions, nodes.length]); 

  // Format edges with styling
  const initialEdges = useMemo(() => edges.map(edge => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
    animated: animationInProgress,
    type: 'smoothstep',
    markerEnd: {
      type: MarkerType.ArrowClosed,
      width: 15,
      height: 15,
      color: '#10b981',
    },
    style: { stroke: "#10b981", strokeWidth: 1.5 }
  })), [edges, animationInProgress]);

  const [reactNodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [reactEdges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Update nodes and edges when props change
  useEffect(() => {
    if (nodes.length === 0) return;
    
    // Use setTimeout to prevent UI blocking
    const timer = setTimeout(() => {
      const updatedNodes = calculateNodePositions();
      
      const updatedEdges = edges.map(edge => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
        animated: animationInProgress,
        type: 'smoothstep',
        markerEnd: {
          type: MarkerType.ArrowClosed,
          width: 15,
          height: 15,
          color: '#10b981',
        },
        style: { stroke: "#10b981", strokeWidth: 1.5 }
      }));
  
      setNodes(updatedNodes);
      setEdges(updatedEdges);
    }, 100);
    
    return () => clearTimeout(timer);
  }, [nodes, edges, animationInProgress, setNodes, setEdges, calculateNodePositions]);

  return (
    <div style={{ width: "100%", height: "700px" }} className="recipe-tree-container border border-gray-700 rounded-lg">
      {isProcessing && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/30 z-10">
          <div className="bg-gray-900 p-4 rounded-lg shadow-lg">
            <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-green-500 mx-auto mb-2"></div>
            <p className="text-green-400">Generating recipe tree visualization...</p>
          </div>
        </div>
      )}
      
      {nodes.length > 0 ? (
        <ReactFlow
          nodes={reactNodes}
          edges={reactEdges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          fitView
          fitViewOptions={{ padding: 0.4, minZoom: 0.1 }}
          minZoom={0.05}
          maxZoom={2}
          attributionPosition="bottom-left"
          proOptions={{ hideAttribution: true }}
        >
          <Controls />
          <MiniMap 
            nodeStrokeColor={(n) => {
              if (n.type === 'target') return '#10b981';
              if (n.type === 'basic') return '#3b82f6';
              if (n.type === 'combining') return '#2563eb';
              return '#6b7280';
            }}
            nodeColor={(n) => {
              if (n.type === 'target') return '#10b981';
              if (n.type === 'basic') return '#3b82f6';
              if (n.type === 'combining') return '#2563eb';
              return '#6b7280';
            }}
            maskColor="#00000080"
          />
          <Background color="#444" gap={16} />
        </ReactFlow>
      ) : (
        <div className="h-full flex items-center justify-center bg-gray-800 rounded-lg">
          <p className="text-gray-400">
            Belum ada data recipe untuk divisualisasikan. Lakukan pencarian untuk menampilkan tree.
          </p>
        </div>
      )}
    </div>
  );
}