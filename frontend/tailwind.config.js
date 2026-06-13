/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: {
          DEFAULT: '#0B0C0E',
          end: '#15171B',
        },
        surface: 'rgba(255,255,255,0.04)',
        hairline: 'rgba(255,255,255,0.08)',
        text: '#F5F6F7',
        muted: '#9AA0A6',
        accent: '#C9A24B',
      },
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
