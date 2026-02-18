import { Link } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';

interface BackButtonProps {
  href?: string;
  label?: string;
}

export const BackButton = ({ href = '/dashboard', label = 'Назад' }: BackButtonProps) => (
  <Link to={href}>
    <button className="flex items-center gap-2 text-slate-500 hover:text-white transition-colors mb-6 group">
      <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
      <span className="text-sm font-medium">{label}</span>
    </button>
  </Link>
);
