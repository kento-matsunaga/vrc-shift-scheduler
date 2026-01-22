import { useInView } from '../../hooks/useInView';

interface FeatureCardProps {
  icon: string;
  title: string;
  description: string;
  delay?: number;
}

export function FeatureCard({ icon, title, description, delay = 0 }: FeatureCardProps) {
  const [ref, isInView] = useInView();

  return (
    <div
      ref={ref}
      className="group relative p-4 sm:p-6 rounded-xl sm:rounded-2xl transition-all duration-500 ease-out"
      style={{
        opacity: isInView ? 1 : 0,
        transform: isInView ? 'translateY(0) scale(1)' : 'translateY(30px) scale(0.95)',
        transitionDelay: `${delay}ms`,
        background: 'linear-gradient(135deg, rgba(79, 70, 229, 0.08) 0%, rgba(139, 92, 246, 0.04) 100%)',
        border: '1px solid rgba(139, 92, 246, 0.15)',
        backdropFilter: 'blur(10px)',
        WebkitBackdropFilter: 'blur(10px)',
      }}
    >
      {/* Glow effect on hover */}
      <div
        className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"
        style={{
          background: 'radial-gradient(circle at 50% 50%, rgba(139, 92, 246, 0.15) 0%, transparent 70%)',
          filter: 'blur(20px)',
        }}
      />

      <div className="relative z-10">
        <div
          className="w-11 h-11 sm:w-14 sm:h-14 rounded-lg sm:rounded-xl flex items-center justify-center text-xl sm:text-2xl mb-3 sm:mb-4 transition-transform duration-300 group-hover:scale-110"
          style={{
            background: 'linear-gradient(135deg, rgba(79, 70, 229, 0.3) 0%, rgba(139, 92, 246, 0.2) 100%)',
            boxShadow: '0 4px 20px rgba(139, 92, 246, 0.2)',
          }}
        >
          {icon}
        </div>
        <h3 className="text-base sm:text-lg font-bold text-white mb-1.5 sm:mb-2">{title}</h3>
        <p className="text-gray-400 text-xs sm:text-sm leading-relaxed">{description}</p>
      </div>
    </div>
  );
}
