import Image from "next/image";

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-green-950 via-green-900 to-green-800 text-white px-6 py-16">
      <div className="max-w-4xl mx-auto bg-gray-900 rounded-xl shadow-lg p-8 border border-green-600">
        <h1 className="text-4xl font-extrabold text-green-400 mb-6 text-center">
          The BrYan Alchemind
        </h1>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold text-green-300 mb-3">
            🧪 Tentang Little Alchemy 2
          </h2>
          <p className="text-lg text-gray-300 leading-relaxed mb-4">
            Little Alchemy 2 adalah permainan berbasis web/aplikasi yang dikembangkan oleh Recloak dan dirilis pada tahun 2017. 
            Tujuan permainan ini adalah membuat 720 elemen berbeda dengan menggabungkan 5 elemen dasar yaitu Air, Earth, Fire, Water, dan Time.
          </p>
          <p className="text-lg text-gray-300 leading-relaxed mb-4">
            Mekanisme permainan ini sangat sederhana, yaitu pemain menggabungkan dua elemen dengan cara drag and drop. 
            Jika kombinasi kedua elemen tersebut valid, maka akan muncul elemen baru. 
            Jika kombinasi tidak valid, maka tidak akan terjadi apa-apa. 
            Permainan ini tersedia di web browser, Android, atau iOS.
          </p>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold text-green-300 mb-3">
            📋 Spesifikasi Tugas Besar
          </h2>
          <p className="text-lg text-gray-300 leading-relaxed mb-4">
            Tugas Besar 2 mata kuliah IF2211 Strategi Algoritma ini bertujuan untuk mengimplementasikan algoritma pencarian
            untuk menyelesaikan permainan Little Alchemy 2. Aplikasi ini mampu menemukan kombinasi elemen 
            yang diperlukan untuk membentuk elemen target dengan menggunakan beberapa algoritma pencarian.
          </p>
          
          <div className="mt-4">
            <h3 className="text-xl font-semibold text-green-200 mb-2">
              ✨ Fitur Utama
            </h3>
            <ul className="list-disc list-inside space-y-2 text-gray-200 pl-2">
              <li>
                Pencarian kombinasi elemen menggunakan algoritma BFS, DFS, dan Bidirectional
              </li>
              <li>
                Mode multiple recipe dengan optimasi multithreading
              </li>
              <li>
                Visualisasi pohon hasil pencarian yang interaktif
              </li>
              <li>
                Menampilkan waktu pencarian dan jumlah node yang dikunjungi
              </li>
            </ul>
          </div>
        </section>

        <section className="mb-8">
          <h2 className="text-2xl font-semibold text-green-300 mb-3">
            🛠️ Teknologi yang Digunakan
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="bg-black/30 p-4 rounded-lg">
              <h3 className="font-semibold text-green-200 mb-2">Frontend</h3>
              <ul className="list-disc list-inside space-y-1 text-gray-300 pl-2">
                <li>Next.js 15.3.2 (React 19)</li>
                <li>TypeScript</li>
                <li>Tailwind CSS 4</li>
                <li>React Flow untuk visualisasi tree</li>
              </ul>
            </div>
            <div className="bg-black/30 p-4 rounded-lg">
              <h3 className="font-semibold text-green-200 mb-2">Backend</h3>
              <ul className="list-disc list-inside space-y-1 text-gray-300 pl-2">
                <li>Golang</li>
                <li>Multithreading untuk pencarian multiple recipe</li>
                <li>REST API untuk komunikasi dengan frontend</li>
              </ul>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-green-300 mb-3">
            👨‍💻 Tim Pengembang - TrioKwekKwek
          </h2>
          <div className="bg-black/30 p-4 rounded-lg">
            <ul className="list-disc list-inside space-y-2 text-gray-200 pl-2">
              <li>Azadi Azhrah (12823024)</li>
              <li>Sebastian Enrico Nathanael (13523134)</li>
              <li>Bryan P. Hutagalung (18222130)</li>
            </ul>
          </div>
        </section>
      </div>
    </div>
  );
}
