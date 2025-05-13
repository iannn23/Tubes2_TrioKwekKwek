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

export interface RecipeTreeProps {
  nodes: Array<{
    id: string;
    label: string;
    type: string;
    tier: number;
    imageUrl?: string;
  }>;
  edges: Array<{
    id: string;
    source: string;
    target: string;
  }>;
  animationInProgress: boolean;
}

// Node styling components
const nodeTypes = {
  target: ({ data }: { data: any }) => (
    <div className="p-3 rounded-lg border-2 border-green-600 bg-green-900/70 text-center min-w-[140px]">
      <div className="font-bold text-green-100">{data.label}</div>
      <div className="text-xs text-green-300">Target (T{data.tier})</div>
    </div>
  ),
  basic: ({ data }: { data: any }) => (
    <div className="p-3 rounded-lg border-2 border-blue-400 bg-blue-900/70 text-center min-w-[140px]">
      <div className="font-bold text-blue-100">{data.label}</div>
      <div className="text-xs text-blue-300">Basic (T{data.tier})</div>
    </div>
  ),
  ingredient: ({ data }: { data: any }) => (
    <div className="p-3 rounded-lg border border-gray-400 bg-gray-800/80 text-center min-w-[140px]">
      <div className="font-semibold text-white">{data.label}</div>
      <div className="text-xs text-gray-300">T{data.tier}</div>
      {data.isCycle && (
        <div className="text-xs text-amber-400 mt-1">(cycle detected)</div>
      )}
    </div>
  ),
  combining: ({ data }: { data: any }) => (
    <div className="p-2 rounded-lg border border-blue-400/30 bg-blue-950/30 text-center min-w-[100px]">
      <div className="text-xs text-blue-300">Combining:</div>
    </div>
  ),
};

export default function RecipeTree({ 
  nodes = [], 
  edges = [], 
  animationInProgress = false 
}: RecipeTreeProps) {
  const [isProcessing, setIsProcessing] = useState(false);
  let nodeIdCounter = 0;
  
  // Convert input nodes to ReactFlow format
  const convertToReactFlowNodes = useCallback((inputNodes) => {
    return inputNodes.map(node => ({
      id: node.id,
      type: node.type === "basic" ? "basic" : 
            node.type === "target" ? "target" : "ingredient",
      // IMPORTANT: Make sure data has the label property properly set
      data: { 
        label: node.label || "Unknown",
        tier: node.tier || 0,
        imageUrl: node.imageUrl
      },
      position: { x: 0, y: 0 },
      sourcePosition: Position.Bottom,
      targetPosition: Position.Top
    }));
  }, []);
  
  const calculateLayout = useCallback(() => {
    if (!nodes.length) return { formattedNodes: [], formattedEdges: [] };
    setIsProcessing(true);
    nodeIdCounter = 0;
    
    try {
      // Convert input nodes to ReactFlow format - ensure labels are properly set
      const reactFlowNodes = convertToReactFlowNodes(nodes);
      
      // Find target node (root of the tree)
      const targetNode = reactFlowNodes.find(node => node.type === "target");
      if (!targetNode) {
        console.warn("No target node found, using fallback layout");
        setIsProcessing(false);
        return { 
          formattedNodes: reactFlowNodes,
          formattedEdges: []
        };
      }
      
      // Create node and edge collections
      const formattedNodes = [];
      const formattedEdges = [];
      
      // Create node lookup for faster access (use label from data)
      const nodeMap = {};
      reactFlowNodes.forEach(node => {
        nodeMap[node.id] = {
          ...node,
          // Ensure the data field has all necessary properties
          data: {
            ...node.data,
            label: node.data.label || "Unknown" // Fallback for missing labels
          }
        };
      });
      
      // Build the child-parent relationships (not parent-child)
      const childParentMap = {};
      edges.forEach(edge => {
        childParentMap[edge.source] = edge.target;
      });
      
      // Build a parent-children map for easier traversal
      const parentChildrenMap = {};
      edges.forEach(edge => {
        if (!parentChildrenMap[edge.target]) {
          parentChildrenMap[edge.target] = [];
        }
        parentChildrenMap[edge.target].push(edge.source);
      });
      
      // Constants for tree layout
      const VERTICAL_SPACING = 80;    // Distance between tiers
      const BASE_HORIZONTAL_SPACING = 160;  // Base horizontal spacing
      
      // Helper function to get node's total descendants (to determine width)
      const countDescendants = (nodeId, visited = new Set()) => {
        if (visited.has(nodeId)) return 0;
        visited.add(nodeId);
        
        const children = parentChildrenMap[nodeId] || [];
        if (children.length === 0) return 1;
        
        let count = 0;
        for (const childId of children) {
          count += countDescendants(childId, visited);
        }
        return Math.max(1, count);
      };
      
      // Create a top-down layout where target node is at top
      const layoutNode = (nodeId, x, y, horizontalSpacing, visited = new Set()) => {
        if (visited.has(nodeId)) {
          // Handle cycle - create a reference node
          const node = nodeMap[nodeId];
          if (!node) return null;
          
          const uniqueNodeId = `cycle_${nodeId}_${nodeIdCounter++}`;
          formattedNodes.push({
            id: uniqueNodeId,
            type: node.type === "basic" ? "basic" : "ingredient",
            data: { 
              label: node.data.label, 
              tier: node.data.tier,
              isCycle: true
            },
            position: { x, y },
            sourcePosition: Position.Bottom,
            targetPosition: Position.Top
          });
          
          return { nodeId: uniqueNodeId, width: horizontalSpacing };
        }
        
        // Add this node to visited
        visited.add(nodeId);
        
        const node = nodeMap[nodeId];
        if (!node) return null;
        
        // Create this node with proper data
        const uniqueNodeId = `${nodeId}_${nodeIdCounter++}`;
        formattedNodes.push({
          id: uniqueNodeId,
          type: node.type,
          data: { 
            label: node.data.label,
            tier: node.data.tier,
            imageUrl: node.data.imageUrl 
          },
          position: { x, y },
          sourcePosition: Position.Bottom,
          targetPosition: Position.Top
        });
        
        // If this node has children, process them
        const children = parentChildrenMap[nodeId] || [];
        if (children.length > 0) {
          // Group children if there are too many (max 2 per level)
          const groupedChildren = [];
          for (let i = 0; i < children.length; i += 2) {
            if (i + 1 < children.length) {
              // Pair of children
              groupedChildren.push([children[i], children[i + 1]]);
            } else {
              // Single child
              groupedChildren.push([children[i]]);
            }
          }
          
          // Calculate positions for the grouped children
          const groupY = y + VERTICAL_SPACING;
          const totalWidth = Math.max(200, groupedChildren.length * 320);
          const startX = x - (totalWidth / 2) + 160;
          
          // Process each group
          groupedChildren.forEach((group, groupIndex) => {
            const groupX = startX + groupIndex * 320;
            
            if (group.length === 1) {
              // Single child - direct connection
              const childId = group[0];
              const result = layoutNode(
                childId, 
                groupX,
                groupY,
                horizontalSpacing * 0.8,
                new Set([...visited])
              );
              
              if (result) {
                // Connect parent to child
                formattedEdges.push({
                  id: `edge_${uniqueNodeId}_to_${result.nodeId}`,
                  source: uniqueNodeId,
                  target: result.nodeId,
                  animated: animationInProgress,
                  type: 'smoothstep'
                });
              }
            } else {
              // Two children - add combining node
              const combiningNodeId = `combining_${uniqueNodeId}_${groupIndex}`;
              formattedNodes.push({
                id: combiningNodeId,
                type: "combining",
                data: { label: "Combining:" },
                position: { x: groupX, y: groupY },
                sourcePosition: Position.Bottom,
                targetPosition: Position.Top
              });
              
              // Connect parent to combining node
              formattedEdges.push({
                id: `edge_${uniqueNodeId}_to_${combiningNodeId}`,
                source: uniqueNodeId,
                target: combiningNodeId,
                animated: animationInProgress,
                type: 'smoothstep'
              });
              
              // Process children
              const childrenY = groupY + 60;
              const leftChildX = groupX - 120;
              const rightChildX = groupX + 120;
              
              // Left child
              const leftChildResult = layoutNode(
                group[0],
                leftChildX,
                childrenY,
                horizontalSpacing * 0.7,
                new Set([...visited])
              );
              
              if (leftChildResult) {
                formattedEdges.push({
                  id: `edge_${combiningNodeId}_to_${leftChildResult.nodeId}`,
                  source: combiningNodeId,
                  target: leftChildResult.nodeId,
                  animated: animationInProgress,
                  type: 'smoothstep'
                });
              }
              
              // Right child
              const rightChildResult = layoutNode(
                group[1],
                rightChildX,
                childrenY,
                horizontalSpacing * 0.7,
                new Set([...visited])
              );
              
              if (rightChildResult) {
                formattedEdges.push({
                  id: `edge_${combiningNodeId}_to_${rightChildResult.nodeId}`,
                  source: combiningNodeId,
                  target: rightChildResult.nodeId,
                  animated: animationInProgress,
                  type: 'smoothstep'
                });
              }
            }
          });
        }
        
        return { nodeId: uniqueNodeId, width: horizontalSpacing };
      };
      
      // Start layout from target node at the top
      // Calculate horizontal spacing based on tree breadth
      let totalNodesCount = countDescendants(targetNode.id);
      let horizontalSpacing = Math.max(BASE_HORIZONTAL_SPACING, 
                                     BASE_HORIZONTAL_SPACING * Math.min(2, Math.sqrt(totalNodesCount / 5)));
      
      // Layout the tree from top to bottom
      layoutNode(targetNode.id, 0, 0, horizontalSpacing);
      
      setIsProcessing(false);
      return { formattedNodes, formattedEdges };
      
    } catch (error) {
      console.error("Error calculating tree layout:", error);
      setIsProcessing(false);
      
      // Simple fallback layout - ensure labels are properly set
      return { 
        formattedNodes: nodes.map((node, index) => ({
          id: node.id || `node_${index}`,
          type: node.type === "basic" ? "basic" : 
                node.type === "target" ? "target" : "ingredient",
          data: { 
            label: node.label || "Unknown",  // Make sure label is set
            tier: node.tier || 0
          },
          position: { x: 0, y: (node.tier || 0) * 120 },
          sourcePosition: Position.Bottom,
          targetPosition: Position.Top
        })),
        formattedEdges: edges.map((edge, index) => ({
          id: edge.id || `edge_${index}`,
          source: edge.source,
          target: edge.target,
          animated: animationInProgress,
          type: 'smoothstep',
          style: { stroke: "#10b981", strokeWidth: 1.5 }
        }))
      };
    }
  }, [nodes, edges, animationInProgress, convertToReactFlowNodes]);

  // Use memo to avoid recalculations
  const { initialNodes, initialEdges } = useMemo(() => {
    if (nodes.length === 0) return { initialNodes: [], initialEdges: [] };
    
    const { formattedNodes, formattedEdges } = calculateLayout();
    
    // Add styling to edges
    const styledEdges = formattedEdges.map(edge => ({
      ...edge,
      markerEnd: {
        type: MarkerType.ArrowClosed,
        width: 12,
        height: 12,
        color: '#10b981',
      },
      style: { stroke: "#10b981", strokeWidth: 1.5 }
    }));
    
    return { 
      initialNodes: formattedNodes,
      initialEdges: styledEdges
    };
  }, [calculateLayout, nodes.length]);

  // React Flow state
  const [reactNodes, setNodes, onNodesChange] = useNodesState(initialNodes || []);
  const [reactEdges, setEdges, onEdgesChange] = useEdgesState(initialEdges || []);

  // Update layout when nodes or edges change
  useEffect(() => {
    if (nodes.length === 0) return;
    
    const timer = setTimeout(() => {
      const { formattedNodes, formattedEdges } = calculateLayout();
      
      // Style the edges
      const styledEdges = formattedEdges.map(edge => ({
        ...edge,
        markerEnd: {
          type: MarkerType.ArrowClosed,
          width: 12,
          height: 12,
          color: '#10b981',
        },
        style: { stroke: "#10b981", strokeWidth: 1.5 }
      }));
  
      setNodes(formattedNodes);
      setEdges(styledEdges);
    }, 100);
    
    return () => clearTimeout(timer);
  }, [nodes, edges, animationInProgress, setNodes, setEdges, calculateLayout]);

  // For debugging only - log nodes and edges
  useEffect(() => {
    if (nodes.length > 0) {
      console.log("Input nodes:", nodes);
    }
  }, [nodes]);

  return (
    <div style={{ width: "100%", height: "800px" }} className="recipe-tree-container border border-gray-700 rounded-lg">
      {isProcessing && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/30 z-10">
          <div className="bg-gray-900 p-4 rounded-lg shadow-lg">
            <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-green-500 mx-auto mb-2"></div>
            <p className="text-green-400">Generating recipe tree...</p>
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
          fitViewOptions={{ 
            padding: 0.3,
            minZoom: 0.05,
            maxZoom: 2
          }}
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
              if (n.type === 'combining') return '#6366f1';
              return '#6b7280';
            }}
            nodeColor={(n) => {
              if (n.type === 'target') return '#10b981';
              if (n.type === 'basic') return '#3b82f6';
              if (n.type === 'combining') return '#6366f1';
              return '#6b7280';
            }}
            maskColor="#00000080"
          />
          <Background color="#444" gap={16} />
        </ReactFlow>
      ) : (
        <div className="h-full flex items-center justify-center bg-gray-800 rounded-lg">
          <p className="text-gray-400">
            No recipe tree data available. Search for an element to visualize its recipe tree.
          </p>
        </div>
      )}
    </div>
  );
}