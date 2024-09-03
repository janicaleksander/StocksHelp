/** @type {import('tailwindcss').Config} */
module.exports = {
  mode:"jit",
  content: [
    "./static/**/*.*",
    "./node_modules/flowbite/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        'c1': '#FDC921',
        'c2': '#FDD85D',
        'c3': '#FFFDF7',
        'c4': '#99D6EA',
        'c5': '#6798C0',
        'c6': '#000000',
        'c7': '#00AB41',
        primary: {"50":"#eff6ff","100":"#dbeafe","200":"#bfdbfe","300":"#93c5fd","400":"#60a5fa","500":"#3b82f6","600":"#2563eb","700":"#1d4ed8","800":"#1e40af","900":"#1e3a8a","950":"#172554"},
      },
      fontFamily: {
        IBM: ["IBM Plex Sans", "sans-serif"],
      }
    },
  },
  plugins: [
    require('flowbite/plugin')({
      charts: true,
    }),
  ],
}
