"use strict";
// state
let currentView = "newScrape";
let scrapes = [];
let results = [];
let uploadedUrls = [];
// DOM elements
const newScrapeBtn = document.getElementById("newScrapeBtn");
const viewResultsBtn = document.getElementById("viewResultsBtn");
const statsBtn = document.getElementById("statsBtn");
const refreshBtn = document.getElementById("refreshBtn");
const newScrapeSection = document.getElementById("newScrapeSection");
const resultsSection = document.getElementById("resultsSection");
const statsSection = document.getElementById("statsSection");
const scrapeForm = document.getElementById("scrapeForm");
const depthInput = document.getElementById("depthInput");
const keywordInput = document.getElementById("keywordInput");
const dropzone = document.getElementById("dropzone");
const fileInput = document.getElementById("fileInput");
const fileInfo = document.getElementById("fileInfo");
const fileName = document.getElementById("fileName");
const fileCount = document.getElementById("fileCount");
const removeFileBtn = document.getElementById("removeFile");
const scrapesList = document.getElementById("scrapesList");
const resultsContainer = document.getElementById("resultsContainer");
const filterUrl = document.getElementById("filterUrl");
const filterTitle = document.getElementById("filterTitle");
const totalScrapes = document.getElementById("totalScrapes");
const totalResults = document.getElementById("totalResults");
const avgDepth = document.getElementById("avgDepth");
// view management
function showView(view) {
    currentView = view;
    newScrapeSection.classList.add("hidden");
    resultsSection.classList.add("hidden");
    statsSection.classList.add("hidden");
    switch (view) {
        case "newScrape":
            newScrapeSection.classList.remove("hidden");
            break;
        case "results":
            resultsSection.classList.remove("hidden");
            loadResults();
            break;
        case "stats":
            statsSection.classList.remove("hidden");
            loadStats();
            break;
    }
}
// API calls
async function fetchScrapes() {
    try {
        const response = await fetch("/api/scrapes");
        if (!response.ok)
            throw new Error("Failed to fetch scrapes");
        scrapes = await response.json();
        renderScrapesList();
    }
    catch (error) {
        console.error("Error fetching scrapes:", error);
        scrapes = [];
        renderScrapesList();
    }
}
async function fetchResults(filters) {
    try {
        const params = new URLSearchParams();
        if (filters?.url)
            params.append("url", filters.url);
        if (filters?.title)
            params.append("title", filters.title);
        const response = await fetch(`/query?${params.toString()}`);
        if (!response.ok)
            throw new Error("Failed to fetch results");
        results = await response.json();
        renderResults();
    }
    catch (error) {
        console.error("Error fetching results:", error);
        results = [];
        renderResults();
    }
}
async function startScrape(url, depth, keyword) {
    try {
        const response = await fetch("/api/scrape", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ url, depth, keyword }),
        });
        if (!response.ok)
            throw new Error("Failed to start scrape");
        alert("Scrape started! Check results in a moment.");
        fetchScrapes();
    }
    catch (error) {
        console.error("Error starting scrape:", error);
        alert("Failed to start scrape. Check console for details.");
    }
}
async function startBulkScrape(urls, depth, keyword) {
    try {
        const response = await fetch("/api/scrape/bulk", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ urls, depth, keyword }),
        });
        if (!response.ok)
            throw new Error("Failed to start bulk scrape");
        alert(`Bulk scrape started for ${urls.length} URLs! Check results in a moment.`);
        fetchScrapes();
    }
    catch (error) {
        console.error("Error starting bulk scrape:", error);
        alert("Failed to start bulk scrape. Check console for details.");
    }
}
// file handling
function handleFileUpload(file) {
    if (!file.name.endsWith('.txt')) {
        alert('Please upload a .txt file');
        return;
    }
    const reader = new FileReader();
    reader.onload = (e) => {
        const content = e.target?.result;
        uploadedUrls = content
            .split('\n')
            .map(line => line.trim())
            .filter(line => line.length > 0 && line.startsWith('http'));
        if (uploadedUrls.length === 0) {
            alert('No valid URLs found in file');
            uploadedUrls = [];
            return;
        }
        // Show file info
        fileName.textContent = file.name;
        fileCount.textContent = `${uploadedUrls.length} URLs loaded`;
        fileInfo.classList.remove('hidden');
        dropzone.querySelector('.dropzone__content')?.classList.add('hidden');
    };
    reader.readAsText(file);
}
function removeFile() {
    uploadedUrls = [];
    fileInput.value = '';
    fileInfo.classList.add('hidden');
    dropzone.querySelector('.dropzone__content')?.classList.remove('hidden');
}
// drag and drop handlers
dropzone.addEventListener('click', () => {
    if (uploadedUrls.length === 0) {
        fileInput.click();
    }
});
dropzone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropzone.classList.add('dragover');
});
dropzone.addEventListener('dragleave', () => {
    dropzone.classList.remove('dragover');
});
dropzone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropzone.classList.remove('dragover');
    const files = e.dataTransfer?.files;
    if (files && files.length > 0) {
        handleFileUpload(files[0]);
    }
});
fileInput.addEventListener('change', (e) => {
    const files = e.target.files;
    if (files && files.length > 0) {
        handleFileUpload(files[0]);
    }
});
removeFileBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    removeFile();
});
// render functions
function renderScrapesList() {
    if (scrapes.length === 0) {
        scrapesList.innerHTML = '<div class="sidebar__empty">No scrapes yet</div>';
        return;
    }
    scrapesList.innerHTML = scrapes
        .sort((a, b) => new Date(b.completed_at).getTime() - new Date(a.completed_at).getTime())
        .map((scrape) => `
            <div class="sidebar__item">
                <div class="scrape-item__url">${truncateUrl(scrape.url)}</div>
                <div class="scrape-item__time">${formatTime(scrape.completed_at)}</div>
            </div>
        `)
        .join("");
}
function renderResults() {
    if (results.length === 0) {
        resultsContainer.innerHTML =
            '<div class="results__empty">No results yet</div>';
        return;
    }
    resultsContainer.innerHTML = results
        .map((result) => `
            <div class="result-card">
                <div class="result-card__url">${result.url}</div>
                <div class="result-card__title">${result.title || "No title"}</div>
                <div class="result-card__price">${result.price || "No price"}</div>
                <div class="result-card__time">${formatTime(result.completed_at)}</div>
            </div>
        `)
        .join("");
}
function loadResults() {
    const urlFilter = filterUrl.value;
    const titleFilter = filterTitle.value;
    fetchResults({ url: urlFilter, title: titleFilter });
}
function loadStats() {
    totalScrapes.textContent = scrapes.length.toString();
    totalResults.textContent = results.length.toString();
    const avgDepthValue = scrapes.length > 0
        ? Math.round(scrapes.reduce((sum, s) => sum + (s.max_pages || 0), 0) /
            scrapes.length)
        : 0;
    avgDepth.textContent = avgDepthValue.toString();
}
// Utility Functions
function truncateUrl(url, maxLength = 30) {
    if (url.length <= maxLength)
        return url;
    return url.substring(0, maxLength) + "...";
}
function formatTime(timestamp) {
    if (!timestamp)
        return "N/A";
    const date = new Date(timestamp);
    return date.toLocaleString();
}
// event listeners
newScrapeBtn.addEventListener("click", () => showView("newScrape"));
viewResultsBtn.addEventListener("click", () => showView("results"));
statsBtn.addEventListener("click", () => showView("stats"));
refreshBtn.addEventListener("click", loadResults);
scrapeForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const depth = parseInt(depthInput.value);
    const keyword = keywordInput.value;
    if (uploadedUrls.length === 0) {
        alert("Please upload a .txt file with URLs");
        return;
    }
    startBulkScrape(uploadedUrls, depth, keyword);
    removeFile();
    depthInput.value = "3";
    keywordInput.value = "price";
});
filterUrl.addEventListener("input", loadResults);
filterTitle.addEventListener("input", loadResults);
// initialize
fetchScrapes();
showView("newScrape");
