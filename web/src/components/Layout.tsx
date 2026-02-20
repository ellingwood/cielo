import { Link, Outlet } from 'react-router-dom';

export default function Layout() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-sky-50 to-indigo-100">
      <nav className="bg-white/80 backdrop-blur border-b border-sky-200 px-6 py-3 flex items-center gap-4 shadow-sm">
        <Link to="/" className="text-xl font-bold text-sky-700 tracking-tight hover:text-sky-900 transition">
          Cielo
        </Link>
        <span className="text-xs text-sky-400 font-medium">AI Agent Orchestration</span>
      </nav>
      <Outlet />
    </div>
  );
}
