/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Japanese-inspired color palette
        sushi: {
          red: '#D32F2F',      // Tuna/Maki
          salmon: '#FF7043',    // Salmon
          wasabi: '#66BB6A',    // Wasabi green
          soy: '#3E2723',       // Soy sauce brown
          rice: '#FFF8E1',      // Rice white
          nori: '#1B1B1B',      // Nori black
          ginger: '#FFB74D',    // Pickled ginger
          wood: '#8D6E63',      // Wooden table
          bamboo: '#AED581',    // Bamboo mat
        },
        primary: '#D32F2F',
        'primary-dark': '#B71C1C',
      },
      fontFamily: {
        japanese: ['Noto Sans JP', 'sans-serif'],
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
        pulseRotate: {
          '0%, 100%': { transform: 'scale(1) rotate(0deg)' },
          '50%': { transform: 'scale(1.1) rotate(3deg)' }
        },
        gentlePulse: {
          '0%, 100%': { transform: 'scale(1)', opacity: '1' },
          '50%': { transform: 'scale(1.05)', opacity: '0.9' }
        },
        slideInFan: {
          'from': { transform: 'translateX(-100%) rotate(-10deg)', opacity: '0' },
          'to': { transform: 'translateX(0) rotate(0deg)', opacity: '1' }
        }
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'fade-out': 'fadeOut 0.3s ease-in-out',
        'pulse-rotate': 'pulseRotate 2s ease-in-out infinite',
        'gentle-pulse': 'gentlePulse 2s ease-in-out infinite',
        'slide-in-fan': 'slideInFan 0.6s ease-out'
      },
      boxShadow: {
        'card': '0 4px 6px rgba(0, 0, 0, 0.1), 0 2px 4px rgba(0, 0, 0, 0.06)',
        'card-hover': '0 10px 15px rgba(0, 0, 0, 0.2), 0 4px 6px rgba(0, 0, 0, 0.1)',
        'japanese': '0 8px 16px rgba(0, 0, 0, 0.15)',
      }
    },
  },
  plugins: [],
}
