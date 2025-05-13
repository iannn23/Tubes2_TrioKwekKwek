"use client";

import { useState, useEffect } from 'react';
// Import elements data directly
import elementsData from '../../../../backend/src/Scraper/elements.json';

interface Element {
  name: string;
  recipes?: any[];
  imageUrl?: string;
}

interface TierData {
  tierNum: number;
  elements: Element[];
}

export default function AboutPage() {
  const [elements, setElements] = useState<TierData[]>([]);
  const [selectedTier, setSelectedTier] = useState(0);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    try {
      // Process the elements data to ensure names are populated
      const processedData = elementsData.map(tier => {
        return {
          tierNum: tier.tierNum,
          elements: tier.elements.map((elem, index) => {
            // If name is missing, use a placeholder with index
            if (!elem.name) {
              return { 
                ...elem,
                name: `Element ${index + 1}` 
              };
            }
            return elem;
          })
        };
      });
      
      setElements(processedData);
    } catch (error) {
      console.error('Error processing elements data:', error);
      // Fallback to empty tiers if there's an error
      setElements([]);
    } finally {
      setLoading(false);
    }
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-950 via-green-900 to-green-800 text-white px-4 py-10">
      <h1 className="text-4xl font-bold text-center mb-10 text-green-400">
        Algoritma yang Digunakan
      </h1>

      <div className="grid grid-cols-1 sm:grid-cols-1 md:grid-cols-1 gap-6 max-w-5xl mx-auto px-2 mb-12">
        <div className="bg-black rounded-lg p-6 shadow-lg border border-green-500 hover:scale-[1.01] transition">
          <h2 className="text-2xl font-bold text-green-300 mb-3">
            🔎 Breadth-First Search (BFS)
          </h2>
          <p className="text-gray-300 mb-4">
            Algoritma BFS melakukan pencarian secara melebar dengan mengeksplorasi semua node pada tingkat kedalaman 
            saat ini sebelum bergerak ke node pada tingkat berikutnya. Di aplikasi ini, BFS mengidentifikasi jalur 
            kombinasi terpendek untuk membuat elemen target.
          </p>
          <p className="text-gray-300 mb-4">
            Implementasi BFS pada aplikasi ini dilakukan dengan langkah-langkah berikut:
          </p>
          <ol className="list-decimal list-inside text-gray-300 mb-4 pl-4 space-y-1">
            <li>Dimulai dengan memasukkan semua elemen dasar ke dalam antrian</li>
            <li>Mengambil elemen dari depan antrian dan mencoba menggunakannya dalam kombinasi yang memungkinkan</li>
            <li>Jika elemen hasil kombinasi belum pernah dikunjungi, tambahkan ke antrian</li>
            <li>Lacak parent dari setiap elemen untuk membangun jalur kembali ke elemen dasar</li>
            <li>Jika target ditemukan, rekonstruksi jalur dari elemen dasar ke target</li>
          </ol>
          <p className="text-gray-300">
            BFS menjamin menemukan jalur terpendek (dengan jumlah kombinasi minimum) karena selalu memeriksa 
            level kedalaman saat ini sebelum pindah ke level yang lebih dalam. Dalam konteks Little Alchemy 2, 
            BFS sangat efektif untuk menemukan resep paling efisien.
          </p>
        </div>
        
        <div className="bg-black rounded-lg p-6 shadow-lg border border-green-600 hover:scale-[1.01] transition mt-6">
          <h2 className="text-2xl font-bold text-green-300 mb-3">
            🌱 Depth-First Search (DFS)
          </h2>
          <p className="text-gray-300 mb-4">
            Algoritma DFS melakukan pencarian secara mendalam, mengeksplorasi satu jalur sampai akhir
            sebelum melakukan backtracking dan mencoba jalur lainnya. Dalam implementasinya untuk Little Alchemy 2,
            DFS menelusuri kemungkinan kombinasi elemen secara mendalam.
          </p>
          <p className="text-gray-300 mb-4">
            Implementasi DFS pada aplikasi ini dilakukan dengan langkah-langkah berikut:
          </p>
          <ol className="list-decimal list-inside text-gray-300 mb-4 pl-4 space-y-1">
            <li>Gunakan elemen dasar sebagai titik awal pencarian</li>
            <li>Eksplor kombinasi elemen secara rekursif hingga kedalaman tertentu</li>
            <li>Batasi kedalaman maksimum untuk menghindari pencarian yang terlalu dalam atau infinite loop</li>
            <li>Hanya pertimbangkan kombinasi yang menghasilkan elemen dengan tier yang lebih tinggi</li>
            <li>Rekam jalur untuk setiap kombinasi valid yang ditemukan</li>
          </ol>
          <p className="text-gray-300">
            DFS memiliki keunggulan dalam hal konsumsi memori yang lebih efisien dibanding BFS, namun tidak 
            menjamin jalur terpendek. Untuk permainan Little Alchemy 2, DFS berguna untuk menemukan beragam 
            jalur alternatif menuju elemen target.
          </p>
        </div>

        <div className="bg-black rounded-lg p-6 shadow-lg border border-green-700 hover:scale-[1.01] transition mt-6">
          <h2 className="text-2xl font-bold text-green-300 mb-3">
            🔁 Bidirectional Search
          </h2>
          <p className="text-gray-300 mb-4">
            Algoritma Bidirectional Search melakukan pencarian dari dua arah secara bersamaan: dari elemen dasar 
            menuju elemen target (forward search), dan dari elemen target kembali ke elemen-elemen yang bisa 
            membentuknya (backward search).
          </p>
          <p className="text-gray-300 mb-4">
            Implementasi Bidirectional Search pada aplikasi ini dilakukan dengan langkah-langkah berikut:
          </p>
          <ol className="list-decimal list-inside text-gray-300 mb-4 pl-4 space-y-1">
            <li>Mulai pencarian forward dari elemen dasar, dan pencarian backward dari elemen target</li>
            <li>Lakukan pencarian berselang-seling antara kedua arah</li>
            <li>Perhatikan hierarki tier untuk menjaga kevalidan kombinasi</li>
            <li>Setiap elemen yang dikunjungi di kedua pencarian menjadi titik pertemuan</li>
            <li>Ketika titik pertemuan ditemukan, gabungkan jalur forward dan backward</li>
          </ol>
          <p className="text-gray-300">
            Bidirectional Search dapat sangat efisien karena mengurangi ruang pencarian secara signifikan dibandingkan 
            pencarian satu arah. Algoritma ini ideal untuk permainan Little Alchemy 2 yang memiliki banyak elemen 
            dengan berbagai jalur kombinasi.
          </p>
        </div>
      </div>

      <div className="max-w-5xl mx-auto mt-12">
        <h2 className="text-3xl font-bold text-center mb-6 text-green-400">
          Daftar Elemen Little Alchemy 2
        </h2>
        
        <div className="mb-4 flex justify-center">
          <select 
            className="bg-gray-800 text-white rounded px-4 py-2 border border-green-500"
            value={selectedTier}
            onChange={(e) => setSelectedTier(Number(e.target.value))}
          >
            {elements.map(tier => (
              <option key={tier.tierNum} value={tier.tierNum}>
                Tier {tier.tierNum} {tier.tierNum === 0 ? '(Elemen Dasar)' : ''}
              </option>
            ))}
          </select>
        </div>

        <div className="bg-black/70 rounded-lg p-4 border border-green-600">
          <h3 className="text-xl font-semibold mb-4 text-green-300">
            Elemen Tier {selectedTier}
          </h3>
          
          {loading ? (
            <p className="text-center text-gray-400 py-4">Memuat data elemen...</p>
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-2">
              {elements
                .find(tier => tier.tierNum === selectedTier)?.elements.map((elem, index) => (
                  <div key={index} className="bg-gray-900 p-2 rounded border border-green-500 text-center">
                    {elem.name}
                  </div>
                )) || (
                  <p className="text-center text-gray-400 col-span-full py-4">
                    Tier ini kosong atau data belum tersedia.
                  </p>
                )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}