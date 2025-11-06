import { getStatusColor, getStatusLabel } from '../../utils/orderHelpers';

/**
 * StatusBadge Component
 * Badge reutilizável para mostrar status de pedidos
 */
const StatusBadge = ({ status, className = '' }) => {
  return (
    <span
      className={`
        inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold border
        ${getStatusColor(status)}
        ${className}
      `}
    >
      {getStatusLabel(status)}
    </span>
  );
};

export default StatusBadge;
