import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Mail, Lock, User, Phone, UserPlus, Loader2, CheckCircle2, XCircle } from 'lucide-react';
import { useAuth } from '../../hooks/useAuth';
import { validateRegisterForm, phoneMask, removePhoneMask, calculatePasswordStrength } from '../../utils/formHelpers';

function Register() {
  const navigate = useNavigate();
  const { register } = useAuth();
  
  // Estado do formulário
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    password: '',
    confirmPassword: '',
  });

  // Estado de loading e erros
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState({});
  const [passwordStrength, setPasswordStrength] = useState({ strength: 0, label: '', color: '', percentage: 0, criteria: {} });

  // Calcular força da senha quando mudar
  useEffect(() => {
    if (formData.password) {
      const strength = calculatePasswordStrength(formData.password);
      setPasswordStrength(strength);
    } else {
      setPasswordStrength({ strength: 0, label: '', color: '', percentage: 0, criteria: {} });
    }
  }, [formData.password]);

  // Manipular mudanças nos inputs
  const handleChange = (e) => {
    const { name, value } = e.target;
    
    // Aplicar máscara no telefone
    if (name === 'phone') {
      const maskedValue = phoneMask(value);
      setFormData(prev => ({
        ...prev,
        [name]: maskedValue,
      }));
    } else {
      setFormData(prev => ({
        ...prev,
        [name]: value,
      }));
    }

    // Limpar erro do campo ao digitar
    if (errors[name]) {
      setErrors(prev => ({
        ...prev,
        [name]: '',
      }));
    }
  };

  // Verificar se formulário é válido
  const isFormValid = () => {
    const validationErrors = validateRegisterForm(formData);
    return Object.keys(validationErrors).length === 0;
  };

  // Manipular submit do formulário
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validar formulário
    const validationErrors = validateRegisterForm(formData);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    // Iniciar loading
    setLoading(true);

    try {
      // Remover máscara do telefone antes de enviar
      const phoneNumbers = removePhoneMask(formData.phone);

      // Chamar função de registro do contexto
      const result = await register(
        formData.email,
        formData.password,
        formData.name,
        phoneNumbers
      );

      if (result.success) {
        // Redirecionar para login após sucesso
        navigate('/login');
      }
    } catch (error) {
      console.error('Erro no registro:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-success-50 to-success-100 flex items-center justify-center px-4 py-12">
      <div className="max-w-md w-full">
        {/* Logo/Header */}
        <div className="text-center mb-8 animate-slide-up">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-success-600 rounded-full mb-4">
            <UserPlus className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Criar Conta
          </h1>
          <p className="text-gray-600">
            Preencha seus dados para começar
          </p>
        </div>

        {/* Formulário */}
        <div className="card animate-slide-up" style={{ animationDelay: '0.1s' }}>
          <form onSubmit={handleSubmit} className="space-y-5">
            {/* Campo Nome */}
            <div>
              <label htmlFor="name" className="label">
                Nome Completo
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <User className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="text"
                  id="name"
                  name="name"
                  value={formData.name}
                  onChange={handleChange}
                  className={`input pl-10 ${errors.name ? 'border-red-500 focus:ring-red-500' : ''}`}
                  placeholder="João Silva"
                  autoComplete="name"
                  disabled={loading}
                />
              </div>
              {errors.name && (
                <p className="mt-1 text-sm text-red-600 animate-fade-in">
                  {errors.name}
                </p>
              )}
            </div>

            {/* Campo Email */}
            <div>
              <label htmlFor="email" className="label">
                Email
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Mail className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="email"
                  id="email"
                  name="email"
                  value={formData.email}
                  onChange={handleChange}
                  className={`input pl-10 ${errors.email ? 'border-red-500 focus:ring-red-500' : ''}`}
                  placeholder="seu@email.com"
                  autoComplete="email"
                  disabled={loading}
                />
              </div>
              {errors.email && (
                <p className="mt-1 text-sm text-red-600 animate-fade-in">
                  {errors.email}
                </p>
              )}
            </div>

            {/* Campo Telefone */}
            <div>
              <label htmlFor="phone" className="label">
                Telefone
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Phone className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="tel"
                  id="phone"
                  name="phone"
                  value={formData.phone}
                  onChange={handleChange}
                  className={`input pl-10 ${errors.phone ? 'border-red-500 focus:ring-red-500' : ''}`}
                  placeholder="(11) 99999-9999"
                  autoComplete="tel"
                  disabled={loading}
                  maxLength={15}
                />
              </div>
              {errors.phone && (
                <p className="mt-1 text-sm text-red-600 animate-fade-in">
                  {errors.phone}
                </p>
              )}
            </div>

            {/* Campo Senha */}
            <div>
              <label htmlFor="password" className="label">
                Senha
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Lock className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="password"
                  id="password"
                  name="password"
                  value={formData.password}
                  onChange={handleChange}
                  className={`input pl-10 ${errors.password ? 'border-red-500 focus:ring-red-500' : ''}`}
                  placeholder="••••••••"
                  autoComplete="new-password"
                  disabled={loading}
                />
              </div>
              {errors.password && (
                <p className="mt-1 text-sm text-red-600 animate-fade-in">
                  {errors.password}
                </p>
              )}

              {/* Indicador de Força da Senha */}
              {formData.password && (
                <div className="mt-2 space-y-2 animate-fade-in">
                  {/* Barra de progresso */}
                  <div className="flex items-center gap-2">
                    <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
                      <div
                        className={`h-full ${passwordStrength.color} transition-all duration-300`}
                        style={{ width: `${passwordStrength.percentage}%` }}
                      />
                    </div>
                    {passwordStrength.label && (
                      <span className="text-xs font-medium text-gray-600">
                        {passwordStrength.label}
                      </span>
                    )}
                  </div>

                  {/* Requisitos da senha */}
                  <div className="text-xs space-y-1">
                    <div className="flex items-center gap-1">
                      {passwordStrength.criteria.length ? (
                        <CheckCircle2 className="w-3 h-3 text-green-600" />
                      ) : (
                        <XCircle className="w-3 h-3 text-gray-400" />
                      )}
                      <span className={passwordStrength.criteria.length ? 'text-green-600' : 'text-gray-500'}>
                        Mínimo 8 caracteres
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      {passwordStrength.criteria.uppercase ? (
                        <CheckCircle2 className="w-3 h-3 text-green-600" />
                      ) : (
                        <XCircle className="w-3 h-3 text-gray-400" />
                      )}
                      <span className={passwordStrength.criteria.uppercase ? 'text-green-600' : 'text-gray-500'}>
                        Letra maiúscula
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      {passwordStrength.criteria.lowercase ? (
                        <CheckCircle2 className="w-3 h-3 text-green-600" />
                      ) : (
                        <XCircle className="w-3 h-3 text-gray-400" />
                      )}
                      <span className={passwordStrength.criteria.lowercase ? 'text-green-600' : 'text-gray-500'}>
                        Letra minúscula
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      {passwordStrength.criteria.numbers ? (
                        <CheckCircle2 className="w-3 h-3 text-green-600" />
                      ) : (
                        <XCircle className="w-3 h-3 text-gray-400" />
                      )}
                      <span className={passwordStrength.criteria.numbers ? 'text-green-600' : 'text-gray-500'}>
                        Número
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      {passwordStrength.criteria.special ? (
                        <CheckCircle2 className="w-3 h-3 text-green-600" />
                      ) : (
                        <XCircle className="w-3 h-3 text-gray-400" />
                      )}
                      <span className={passwordStrength.criteria.special ? 'text-green-600' : 'text-gray-500'}>
                        Caractere especial
                      </span>
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Campo Confirmar Senha */}
            <div>
              <label htmlFor="confirmPassword" className="label">
                Confirmar Senha
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Lock className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="password"
                  id="confirmPassword"
                  name="confirmPassword"
                  value={formData.confirmPassword}
                  onChange={handleChange}
                  className={`input pl-10 ${errors.confirmPassword ? 'border-red-500 focus:ring-red-500' : ''}`}
                  placeholder="••••••••"
                  autoComplete="new-password"
                  disabled={loading}
                />
              </div>
              {errors.confirmPassword && (
                <p className="mt-1 text-sm text-red-600 animate-fade-in">
                  {errors.confirmPassword}
                </p>
              )}
            </div>

            {/* Botão Submit */}
            <button
              type="submit"
              disabled={loading || !isFormValid()}
              className="btn-success w-full flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed mt-6"
            >
              {loading ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  <span>Cadastrando...</span>
                </>
              ) : (
                <>
                  <UserPlus className="w-5 h-5" />
                  <span>Cadastrar</span>
                </>
              )}
            </button>
          </form>

          {/* Link para Login */}
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              Já tem uma conta?{' '}
              <Link
                to="/login"
                className="link"
              >
                Faça login aqui
              </Link>
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-sm text-gray-600 animate-fade-in" style={{ animationDelay: '0.2s' }}>
          <p>© 2025 Pedidos Online. Todos os direitos reservados.</p>
        </div>
      </div>
    </div>
  );
}

export default Register;
