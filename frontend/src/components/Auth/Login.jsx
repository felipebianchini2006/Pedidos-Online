import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Mail, Lock, LogIn, Loader2 } from 'lucide-react';
import { useAuth } from '../../hooks/useAuth';
import { validateLoginForm } from '../../utils/formHelpers';

function Login() {
  const navigate = useNavigate();
  const { login } = useAuth();
  
  // Estado do formulário
  const [formData, setFormData] = useState({
    email: '',
    password: '',
  });

  // Estado de loading e erros
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState({});

  // Manipular mudanças nos inputs
  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));

    // Limpar erro do campo ao digitar
    if (errors[name]) {
      setErrors(prev => ({
        ...prev,
        [name]: '',
      }));
    }
  };

  // Manipular submit do formulário
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validar formulário
    const validationErrors = validateLoginForm(formData);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    // Iniciar loading
    setLoading(true);

    try {
      // Chamar função de login do contexto
      const result = await login(formData.email, formData.password);

      if (result.success) {
        // Redirecionar para home após sucesso
        navigate('/');
      }
    } catch (error) {
      console.error('Erro no login:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-primary-100 flex items-center justify-center px-4 py-12">
      <div className="max-w-md w-full">
        {/* Logo/Header */}
        <div className="text-center mb-8 animate-slide-up">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-full mb-4">
            <LogIn className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Bem-vindo de volta!
          </h1>
          <p className="text-gray-600">
            Faça login para acessar sua conta
          </p>
        </div>

        {/* Formulário */}
        <div className="card animate-slide-up" style={{ animationDelay: '0.1s' }}>
          <form onSubmit={handleSubmit} className="space-y-6">
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
                  autoComplete="current-password"
                  disabled={loading}
                />
              </div>
              {errors.password && (
                <p className="mt-1 text-sm text-red-600 animate-fade-in">
                  {errors.password}
                </p>
              )}
            </div>

            {/* Link Esqueci a senha (futuro) */}
            <div className="flex items-center justify-end">
              <button
                type="button"
                className="text-sm text-primary-600 hover:text-primary-700 font-medium transition-colors"
                disabled
              >
                Esqueceu a senha?
              </button>
            </div>

            {/* Botão Submit */}
            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  <span>Entrando...</span>
                </>
              ) : (
                <>
                  <LogIn className="w-5 h-5" />
                  <span>Entrar</span>
                </>
              )}
            </button>
          </form>

          {/* Link para Registro */}
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              Não tem uma conta?{' '}
              <Link
                to="/register"
                className="link"
              >
                Cadastre-se aqui
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

export default Login;
