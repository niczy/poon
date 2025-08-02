import { FileBrowser } from '@/components/FileBrowser';

export default function Home() {
  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      {/* Header */}
      <header className="bg-gradient-to-r from-blue-600 to-purple-700 text-white shadow-lg">
        <div className="max-w-6xl mx-auto px-4 py-8">
          <div className="flex items-center gap-3 mb-2">
            <span className="text-4xl">📁</span>
            <h1 className="text-4xl font-bold">
              Poon Monorepo Browser
            </h1>
          </div>
          <p className="text-xl opacity-90 font-light">
            Browse and explore files in the Poon monorepo system
          </p>
        </div>
      </header>
      
      {/* Main Content */}
      <main className="flex-1 max-w-6xl mx-auto w-full p-6">
        <FileBrowser initialPath="/" />
      </main>
      
      {/* Footer */}
      <footer className="bg-gray-800 text-gray-300 text-center py-4">
        <p>Powered by Poon • Next.js + TypeScript + Tailwind CSS</p>
      </footer>
    </div>
  );
}