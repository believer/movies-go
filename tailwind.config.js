/** @type {import('tailwindcss').Config} */
module.exports = {
	content: [
		"./views/**/*.html",
		"./views/**/*.templ",
		"./components/**/*.templ",
		"./**/*.go",
	],
	theme: {
		extend: {},
	},
	plugins: [],
};
