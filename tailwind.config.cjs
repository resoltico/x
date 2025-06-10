module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./app/tailwind.css",
    "./components/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f5faff',
          100: '#e0f0ff',
          200: '#b8dcff',
          300: '#8fc8ff',
          400: '#66b4ff',
          500: '#3da0ff',
          600: '#148cff',
          700: '#006cd6',
          800: '#004fa3',
          900: '#003370',
        },
      },
    },
  },
  plugins: [],
};
