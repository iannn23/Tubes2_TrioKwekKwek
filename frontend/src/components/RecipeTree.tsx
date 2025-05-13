import React, { useEffect, useState } from "react";
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  Node,
  Edge,
} from "reactflow";
import dagre from "dagre";
import "reactflow/dist/style.css";

interface RecipeTreeProps {
  nodes: Node[];
  edges: Edge[];
  animationInProgress?: boolean;
}

// Initialize dagre graph for layout
const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));
const NODE_WIDTH = 180;
const NODE_HEIGHT = 60;

/**
 * Applies a top-to-bottom tree layout to nodes and edges
 */
const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
  dagreGraph.setGraph({ rankdir: "TB" });

  // Set nodes with dimensions
  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
  });

  // Set edges
  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  // Perform layout
  dagre.layout(dagreGraph);

  // Apply computed positions
  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - NODE_WIDTH / 2,
        y: nodeWithPosition.y - NODE_HEIGHT / 2,
      },
      // Fix position so React Flow doesn't override
      targetPosition: "top",
      sourcePosition: "bottom",
    };
  });

  return { nodes: layoutedNodes, edges };
};

const RecipeTree: React.FC<RecipeTreeProps> = ({ nodes, edges }) => {
  const [layoutedNodes, setLayoutedNodes] = useState<Node[]>([]);
  const [layoutedEdges, setLayoutedEdges] = useState<Edge[]>([]);

  useEffect(() => {
    if (nodes.length === 0 || edges.length === 0) {
      setLayoutedNodes(nodes);
      setLayoutedEdges(edges);
      return;
    }

    const { nodes: ln, edges: le } = getLayoutedElements(nodes, edges);
    setLayoutedNodes(ln);
    setLayoutedEdges(le);
  }, [nodes, edges]);

  return (
    <div style={{ width: "100%", height: "600px" }}>
      <ReactFlow
        nodes={layoutedNodes}
        edges={layoutedEdges}
        fitView
        attributionPosition="bottom-left"
      >
        <Background gap={12} />
        <Controls />
        <MiniMap
          nodeColor={(node) => (node.type === "target" ? "#10B981" : "#3B82F6")}
        />
      </ReactFlow>
    </div>
  );
};

export default RecipeTree;
