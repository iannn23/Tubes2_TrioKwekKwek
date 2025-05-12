"use client";

export default function AboutPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-green-950 via-green-900 to-green-800 text-white px-4 py-10">
      <h1 className="text-4xl font-bold text-center mb-10 text-green-400">
        Algoritma yang Digunakan
      </h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6 max-w-7xl mx-auto px-2">
        <div className="bg-black rounded-lg p-6 shadow-lg border border-green-500 hover:scale-[1.02] transition">
          <h2 className="text-2xl font-bold text-green-300 mb-3">
            🔎 Breadth-First Search (BFS)
          </h2>
          <p className="text-sm text-gray-300">ini bfs</p>
        </div>
        <div className="bg-black rounded-lg p-6 shadow-lg border border-green-600 hover:scale-[1.02] transition">
          <h2 className="text-2xl font-bold text-green-300 mb-3">
            🌱 Depth-First Search (DFS)
          </h2>
          <p className="text-sm text-gray-300">ini dfs</p>
        </div>
        <div className="bg-black rounded-lg p-6 shadow-lg border border-green-700 hover:scale-[1.02] transition">
          <h2 className="text-2xl font-bold text-green-300 mb-3">
            🔁 Bidirectional Search
          </h2>
          <p className="text-sm text-gray-300">ini Bidirectional</p>
        </div>
      </div>


    </div>
  );
}
