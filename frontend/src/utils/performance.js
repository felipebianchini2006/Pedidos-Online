/**
 * Performance Utilities
 * Funções para otimização de performance
 */

import { useState, useEffect } from 'react';

/**
 * Debounce - Atrasa a execução de uma função até que ela pare de ser chamada por um período
 * Útil para inputs de busca, resize events, etc.
 * 
 * @param {Function} func - Função a ser debounced
 * @param {number} wait - Tempo de espera em ms
 * @returns {Function} Função debounced
 * 
 * @example
 * const handleSearch = debounce((value) => {
 *   fetchResults(value);
 * }, 500);
 */
export const debounce = (func, wait = 300) => {
  let timeout;
  
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};

/**
 * Throttle - Limita a frequência de execução de uma função
 * Útil para scroll events, mouse move, etc.
 * 
 * @param {Function} func - Função a ser throttled
 * @param {number} limit - Tempo mínimo entre execuções em ms
 * @returns {Function} Função throttled
 * 
 * @example
 * const handleScroll = throttle(() => {
 *   console.log('Scrolling');
 * }, 200);
 */
export const throttle = (func, limit = 200) => {
  let inThrottle;
  
  return function(...args) {
    if (!inThrottle) {
      func.apply(this, args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
};

/**
 * Lazy Load Image - Carrega imagem apenas quando visível
 * 
 * @param {string} src - URL da imagem
 * @param {string} placeholder - URL da imagem placeholder
 * @returns {Object} Estado da imagem
 * 
 * @example
 * const { src, loading } = useLazyImage(imageUrl, placeholderUrl);
 */
export const useLazyImage = (src, placeholder = '') => {
  const [imageSrc, setImageSrc] = useState(placeholder);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    const img = new Image();
    img.src = src;
    
    img.onload = () => {
      setImageSrc(src);
      setLoading(false);
    };
    
    img.onerror = () => {
      setImageSrc(placeholder);
      setLoading(false);
    };
  }, [src, placeholder]);
  
  return { src: imageSrc, loading };
};

/**
 * Memoize - Cache de resultados de funções
 * 
 * @param {Function} fn - Função a ser memoizada
 * @returns {Function} Função memoizada
 * 
 * @example
 * const expensiveCalculation = memoize((n) => {
 *   return n * 2;
 * });
 */
export const memoize = (fn) => {
  const cache = new Map();
  
  return (...args) => {
    const key = JSON.stringify(args);
    
    if (cache.has(key)) {
      return cache.get(key);
    }
    
    const result = fn(...args);
    cache.set(key, result);
    return result;
  };
};

/**
 * Batch Updates - Agrupa múltiplas atualizações
 * 
 * @param {Function} fn - Função a ser executada
 * @param {number} delay - Delay entre execuções
 * @returns {Function} Função batched
 */
export const batchUpdates = (fn, delay = 100) => {
  let pending = [];
  let timeoutId = null;
  
  return (...args) => {
    pending.push(args);
    
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    
    timeoutId = setTimeout(() => {
      fn(pending);
      pending = [];
      timeoutId = null;
    }, delay);
  };
};

/**
 * Request Animation Frame wrapper para animações suaves
 * 
 * @param {Function} callback - Função a ser executada
 * @returns {number} Request ID
 */
export const smoothAnimation = (callback) => {
  let ticking = false;
  
  return (...args) => {
    if (!ticking) {
      requestAnimationFrame(() => {
        callback(...args);
        ticking = false;
      });
      ticking = true;
    }
  };
};

/**
 * Intersection Observer Hook para lazy loading
 * 
 * @param {Object} options - Opções do Intersection Observer
 * @returns {Array} [ref, isIntersecting]
 * 
 * @example
 * const [ref, isVisible] = useIntersectionObserver({ threshold: 0.5 });
 */
export const useIntersectionObserver = (options = {}) => {
  const [isIntersecting, setIsIntersecting] = useState(false);
  const [node, setNode] = useState(null);
  
  useEffect(() => {
    if (!node) return;
    
    const observer = new IntersectionObserver(([entry]) => {
      setIsIntersecting(entry.isIntersecting);
    }, options);
    
    observer.observe(node);
    
    return () => {
      observer.disconnect();
    };
  }, [node, options]);
  
  return [setNode, isIntersecting];
};

/**
 * Local Storage com cache
 */
export const cachedLocalStorage = {
  cache: new Map(),
  
  getItem(key) {
    if (this.cache.has(key)) {
      return this.cache.get(key);
    }
    
    const item = localStorage.getItem(key);
    this.cache.set(key, item);
    return item;
  },
  
  setItem(key, value) {
    this.cache.set(key, value);
    localStorage.setItem(key, value);
  },
  
  removeItem(key) {
    this.cache.delete(key);
    localStorage.removeItem(key);
  },
  
  clear() {
    this.cache.clear();
    localStorage.clear();
  }
};

export default {
  debounce,
  throttle,
  useLazyImage,
  memoize,
  batchUpdates,
  smoothAnimation,
  useIntersectionObserver,
  cachedLocalStorage,
};
