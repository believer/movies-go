const series = document.getElementById("series");

// Display series name below series' input when selected
series.addEventListener("change", (el) => {
	const value = el.currentTarget.value;
	const seriesList = document.getElementById("series_list");
	const seriesName = document.getElementById("series_name");

	if (value === "") {
		seriesName.classList.add("hidden");

		return;
	}

	const option = seriesList.querySelector(`option[value="${value}"]`);

	seriesName.innerText = option.getAttribute("label");
	seriesName.classList.remove("hidden");
});
