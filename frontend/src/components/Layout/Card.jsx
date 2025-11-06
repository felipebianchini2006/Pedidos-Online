/**
 * Card Component - Componente de card reutilizável
 * 
 * Um componente flexível para criar cards com diferentes variantes e tamanhos
 */

const Card = ({ 
  children, 
  title, 
  subtitle,
  footer,
  variant = 'default',
  padding = 'normal',
  hover = false,
  className = '' 
}) => {
  const variantClasses = {
    default: 'bg-white border border-gray-200',
    outlined: 'bg-white border-2 border-gray-300',
    elevated: 'bg-white shadow-lg',
    ghost: 'bg-transparent',
    primary: 'bg-gradient-to-br from-blue-50 to-indigo-50 border border-blue-200',
    success: 'bg-green-50 border border-green-200',
    warning: 'bg-yellow-50 border border-yellow-200',
    danger: 'bg-red-50 border border-red-200',
  };

  const paddingClasses = {
    none: 'p-0',
    small: 'p-4',
    normal: 'p-6',
    large: 'p-8',
  };

  const hoverClass = hover ? 'hover:shadow-xl hover:-translate-y-1 transition-all duration-300 cursor-pointer' : '';

  return (
    <div 
      className={`
        rounded-lg 
        ${variantClasses[variant]} 
        ${paddingClasses[padding]} 
        ${hoverClass}
        ${className}
      `}
    >
      {/* Header */}
      {(title || subtitle) && (
        <div className="mb-4 pb-4 border-b border-gray-200">
          {title && (
            <h3 className="text-xl font-bold text-gray-800 mb-1">
              {title}
            </h3>
          )}
          {subtitle && (
            <p className="text-sm text-gray-600">
              {subtitle}
            </p>
          )}
        </div>
      )}

      {/* Content */}
      <div className="card-content">
        {children}
      </div>

      {/* Footer */}
      {footer && (
        <div className="mt-4 pt-4 border-t border-gray-200">
          {footer}
        </div>
      )}
    </div>
  );
};

export default Card;
