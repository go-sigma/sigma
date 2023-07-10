/** @type {import("tailwindcss").Config} */
module.exports = {
  darkMode: "class",
  content: ["./src/**/*.{html,js,tsx,scss}", "./index.html"],
  theme: {
    extend: {},
  },
  plugins: [
    require("@tailwindcss/typography"),
    require("@tailwindcss/forms"),
    require("@tailwindcss/aspect-ratio"),
  ],
}
