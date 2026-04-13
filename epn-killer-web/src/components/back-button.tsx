import { useNavigate } from 'react-router-dom';
import { ChevronLeft } from 'lucide-react';

interface BackButtonProps {
  href?: string;
  label?: string;
}

export const BackButton = ({ href, label = 'Назад' }: BackButtonProps) => {
  const navigate = useNavigate();

  const handleClick = () => {
    if (href) {
      navigate(href);
    } else if (window.history.length > 1) {
      navigate(-1);
    } else {
      navigate('/dashboard');
    }
  };

  return (
    <button
      onClick={handleClick}
      className="flex items-center gap-1 text-slate-500 hover:text-white transition-colors mb-6 group"
    >
      <ChevronLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
      <span className="text-sm font-medium">{label}</span>
    </button>
  );
};
