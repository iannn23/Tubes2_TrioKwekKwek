import Image from "next/image";
import Header from "../components/Header";

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-green-950 via-green-900 to-green-800 text-white px-6 py-16">
      <div className="max-w-4xl mx-auto bg-gray-900 rounded-xl shadow-lg p-8 border border-green-600">
        <h1 className="text-4xl font-extrabold text-green-400 mb-6 text-center">
          Tentang Aplikasi
        </h1>

        <p className="text-lg text-gray-300 leading-relaxed mb-6">
          <strong className="text-white">The BrYan Alchemind</strong> sebuah aplikasi eksploratif yang membantu pengguna menemukan kombinasi resep dari permainan Little Alchemy 2 dengan visualisasi tree dari komponen . Aplikasi ini dibangun oleh{" "}
          <span className="text-green-300 font-semibold">Bryan</span> dan{" "}
          <span className="text-green-300 font-semibold">Sebastian</span> sebagai
          bagian dari Tugas Besar mata kuliah{" "}
          <span className="text-blue-400">IF2211 Strategi Algoritma</span>.
        </p>

        <div className="mt-4">
          <h2 className="text-2xl font-semibold text-green-300 mb-3">
            ✨ Fitur Utama
          </h2>
          <ul className="list-disc list-inside space-y-2 text-gray-200 text-sm pl-2">
            <li>
              <span className="text-white">
                Pencarian recipe menggunakan algoritma BFS, DFS, dan Bidirectional
              </span>
            </li>
            <li>
              <span className="text-white">
                Mode multiple recipe (multithreading)
              </span>
            </li>
            <li>
              <span className="text-white">
                Visualisasi pohon hasil pencarian yang interaktif
              </span>
            </li>
            <li>
              <span className="text-white">
                Interface modern dan responsif dengan Next.js dan Tailwind
              </span>
            </li>
            <li>
              <span className="text-white">
                Didukung oleh bahasa Go sebagai backend
              </span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
