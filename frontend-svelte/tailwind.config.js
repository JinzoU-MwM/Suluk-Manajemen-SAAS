/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{html,js,svelte,ts}"
  ],
  theme: {
    extend: {
      colors: {
        // Suluk brand — Primary Green #1B7F5A (interactive), deep #0F3D2E, mint #E8F4EF
        primary: {
          50: '#E8F4EF',
          100: '#d3ebe1',
          200: '#a7d6c2',
          300: '#74bb9f',
          400: '#3f9d79',
          500: '#22865f',
          600: '#1B7F5A',
          700: '#155f44',
          800: '#0F3D2E',
          900: '#0a2a1f',
        },
        emerald: {
          50: '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          300: '#6ee7b7',
          400: '#34d399',
          500: '#10b981',
          600: '#059669',
          700: '#047857',
        },
        // Suluk Golden #C99A2E
        gold: {
          400: '#e0bb5e',
          500: '#C99A2E',
          600: '#a87f22',
        },
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'system-ui', 'sans-serif'],
        serif: ['"Playfair Display"', 'Georgia', 'serif'],
        display: ['"Playfair Display"', 'Georgia', 'serif'],
      },
      animation: {
        float: 'float 6s ease-in-out infinite',
        'float-delay': 'float 6s ease-in-out 2s infinite',
        'pulse-slow': 'pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'fade-up': 'fadeUp 0.8s ease-out forwards',
        'fade-up-delay-1': 'fadeUp 0.8s ease-out 0.1s forwards',
        'fade-up-delay-2': 'fadeUp 0.8s ease-out 0.2s forwards',
        'fade-up-delay-3': 'fadeUp 0.8s ease-out 0.3s forwards',
        'fade-up-delay-4': 'fadeUp 0.8s ease-out 0.4s forwards',
        'slide-right': 'slideRight 0.8s ease-out forwards',
        'spin-slow': 'spin 30s linear infinite',
      },
      keyframes: {
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-20px)' },
        },
        fadeUp: {
          '0%': { opacity: '0', transform: 'translateY(30px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        slideRight: {
          '0%': { opacity: '0', transform: 'translateX(-30px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' },
        },
      },
    },
  },
  plugins: [],
}
