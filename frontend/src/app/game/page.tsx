"use client";

import RecipeTree from "@/components/RecipeTree";
import { useState } from "react";

export default function GamePage() {
  const [algorithm, setAlgorithm] = useState("BFS");
  const [mode, setMode] = useState<"single" | "multiple">("single");
  const [maxRecipes, setMaxRecipes] = useState(3);

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-950 via-green-900 to-green-800 text-white p-6">
      <h1 className="text-4xl font-bold mb-6 text-green-400">Game Page</h1>

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
            {algo === "BFS"} {algo === "DFS"} {algo === "BIDIRECTIONAL"} {algo}
          </button>
        ))}
      </div>

      <p className="text-lg mb-10">
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
          🔹 Satu Recipe (Bebas)
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
            value={maxRecipes}
            onChange={(e) => setMaxRecipes(Number(e.target.value))}
            className="bg-gray-900 border border-green-500 text-white rounded-md px-3 py-2"
          />
        </div>
      )}

      {/* Ringkasan */}
      <p className="text-lg mb-10">
        Mode:{" "}
        <span className="text-green-400 font-semibold">
          {mode === "single"
            ? "Satu Recipe Bebas"
            : `Multiple Recipe (max ${maxRecipes})`}
        </span>{" "}
        | Algoritma:{" "}
        <span className="text-green-400 font-semibold">{algorithm}</span>
      </p>

      <div className="bg-gray-900 border border-green-500 p-6 rounded-lg shadow-lg">
        <h2 className="text-2xl font-semibold text-green-300 mb-4">
          Visualisasi Tree Elements Recipe
        </h2>
        <RecipeTree />
      </div>
    </div>
  );
}
