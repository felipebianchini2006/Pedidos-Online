import { getStatusLabel, getStatusIcon } from '../../utils/orderHelpers';

/**
 * OrderTimeline Component
 * Timeline vertical mostrando o histórico de status do pedido
 */
const OrderTimeline = ({ currentStatus }) => {
  const steps = [
    { status: 'pending', label: 'Pedido Criado' },
    { status: 'confirmed', label: 'Confirmado' },
    { status: 'preparing', label: 'Em Preparação' },
    { status: 'shipped', label: 'Enviado' },
    { status: 'delivered', label: 'Entregue' },
  ];

  const statusOrder = ['pending', 'confirmed', 'preparing', 'shipped', 'delivered', 'cancelled'];
  const currentIndex = statusOrder.indexOf(currentStatus);

  const getStepState = (index, stepStatus) => {
    if (currentStatus === 'cancelled') {
      return 'cancelled';
    }
    
    const stepIndex = statusOrder.indexOf(stepStatus);
    if (stepIndex < currentIndex) return 'completed';
    if (stepIndex === currentIndex) return 'current';
    return 'upcoming';
  };

  const getStepColor = (state) => {
    switch (state) {
      case 'completed':
        return {
          bg: 'bg-green-500',
          text: 'text-green-600',
          border: 'border-green-500',
          line: 'bg-green-500',
        };
      case 'current':
        return {
          bg: 'bg-blue-500',
          text: 'text-blue-600',
          border: 'border-blue-500',
          line: 'bg-gray-300',
        };
      case 'cancelled':
        return {
          bg: 'bg-red-500',
          text: 'text-red-600',
          border: 'border-red-500',
          line: 'bg-red-300',
        };
      default:
        return {
          bg: 'bg-gray-300',
          text: 'text-gray-400',
          border: 'border-gray-300',
          line: 'bg-gray-300',
        };
    }
  };

  if (currentStatus === 'cancelled') {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6">
        <div className="flex items-center space-x-3">
          <div className="w-12 h-12 bg-red-500 rounded-full flex items-center justify-center">
            <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              {getStatusIcon('cancelled')}
            </svg>
          </div>
          <div>
            <p className="font-semibold text-red-800 text-lg">Pedido Cancelado</p>
            <p className="text-sm text-red-600">Este pedido foi cancelado.</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flow-root">
      <ul className="-mb-8">
        {steps.map((step, index) => {
          const state = getStepState(index, step.status);
          const colors = getStepColor(state);
          const isLast = index === steps.length - 1;

          return (
            <li key={step.status}>
              <div className="relative pb-8">
                {!isLast && (
                  <span
                    className={`absolute left-6 top-6 -ml-px h-full w-0.5 ${colors.line}`}
                    aria-hidden="true"
                  />
                )}
                <div className="relative flex items-start space-x-4">
                  <div>
                    <span
                      className={`
                        h-12 w-12 rounded-full flex items-center justify-center ring-8 ring-white
                        ${colors.bg}
                        ${state === 'current' ? 'animate-pulse' : ''}
                      `}
                    >
                      <svg
                        className="h-6 w-6 text-white"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        {getStatusIcon(step.status)}
                      </svg>
                    </span>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div>
                      <p
                        className={`
                          text-base font-semibold
                          ${state === 'upcoming' ? 'text-gray-400' : colors.text}
                        `}
                      >
                        {step.label}
                      </p>
                      <p
                        className={`
                          text-sm mt-1
                          ${state === 'upcoming' ? 'text-gray-400' : 'text-gray-600'}
                        `}
                      >
                        {getStatusLabel(step.status)}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </li>
          );
        })}
      </ul>
    </div>
  );
};

export default OrderTimeline;
