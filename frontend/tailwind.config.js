/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#667eea',
        'primary-dark': '#5568d3',
      },
      keyframes: {
        fadeIn: {
          'from': { opacity: '0', transform: 'translateY(20px)' },
          'to': { opacity: '1', transform: 'translateY(0)' }
        },
        fadeOut: {
          'from': { opacity: '1', transform: 'translateY(0)' },
          'to': { opacity: '0', transform: 'translateY(-20px)' }
        },
        pulse: {
          '0%, 100%': { transform: 'scale(1)', boxShadow: '0 2px 4px rgba(0,0,0,0.1)' },
          '50%': { transform: 'scale(1.05)', boxShadow: '0 4px 8px rgba(76, 175, 80, 0.4)' }
        }
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'fade-out': 'fadeOut 0.3s ease-in-out',
        'pulse': 'pulse 2s infinite'
      }
    },
  },
  plugins: [],
}
