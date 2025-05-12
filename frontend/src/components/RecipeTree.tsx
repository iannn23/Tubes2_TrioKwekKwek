"use client";

import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Node,
} from "reactflow";
import "reactflow/dist/style.css";

const initialNodes: Node[] = [
  {
    id: "1",
    position: { x: 0, y: 0 },
    data: { label: "Brick" },
    type: "default",
  },
  {
    id: "2",
    position: { x: -100, y: 100 },
    data: { label: "Mud" },
  },
  {
    id: "3",
    position: { x: 100, y: 100 },
    data: { label: "Fire" },
  },
  {
    id: "4",
    position: { x: -150, y: 200 },
    data: { label: "Water" },
  },
  {
    id: "5",
    position: { x: -50, y: 200 },
    data: { label: "Earth" },
  },
];

const initialEdges: Edge[] = [
  { id: "e1-2", source: "2", target: "1", animated: true },
  { id: "e3-1", source: "3", target: "1", animated: true },
  { id: "e4-2", source: "4", target: "2" },
  { id: "e5-2", source: "5", target: "2" },
];

export default function RecipeTree() {
  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  const onConnect = (connection: Connection) =>
    setEdges((eds) => addEdge(connection, eds));



  return (
    <div style={{ width: "100%", height: "600px" }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        fitView
      >
        <Controls />
        <MiniMap />
        <Background />
      </ReactFlow>
    </div>
  );
}
