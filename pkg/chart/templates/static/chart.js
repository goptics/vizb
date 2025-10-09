// Store original chart data for each chart instance
const originalChartData = new Map();

// Sort chart with per-workload independent sorting
function sortChart(chartInstance, sortOrder) {
    // Get chart ID to retrieve original data
    const chartDom = chartInstance.getDom();
    const chartId = chartDom.id;

    // Store original data if not already stored
    if (!originalChartData.has(chartId)) {
        const option = chartInstance.getOption();
        originalChartData.set(chartId, JSON.parse(JSON.stringify(option)));
    }

    // Always work from the original data
    const originalData = originalChartData.get(chartId);
    if (!originalData.series || originalData.series.length === 0) {
        return;
    }

    if (sortOrder === 'default') {
        // Restore original structure
        chartInstance.clear();
        chartInstance.setOption(JSON.parse(JSON.stringify(originalData)));
        return;
    }

    const workloads = originalData.xAxis[0].data;
    const subjects = originalData.series.map(function(s) { return s.name; });

    // Build data map: data[workload][subject] = value
    const dataMap = {};
    workloads.forEach(function(workload, wIdx) {
        dataMap[workload] = {};
        originalData.series.forEach(function(series) {
            const dataPoint = series.data[wIdx];
            const value = typeof dataPoint === 'object' ? parseFloat(dataPoint.value) : parseFloat(dataPoint);
            dataMap[workload][series.name] = isNaN(value) ? 0 : value;
        });
    });

    // For each workload, sort subjects by their values
    const subjectMap = new Map()

    workloads.forEach(function(workload) {
        for (const subject of subjects) {
            const subjectValue = dataMap[workload][subject]

            if (subjectMap.has(subject)) {
                subjectMap.set(subject, {
                    subject,
                    total: subjectMap.get(subject).total + subjectValue
                })
            } else {
                subjectMap.set(subject, {
                    subject,
                    total: subjectValue
                })
            }
        }
    });

    const orderSubject = Array.from(subjectMap.values())
        .sort((a, b) => {
            if (sortOrder === 'asc') {
                return a.total - b.total;
            }

            return b.total - a.total;
        })
        .map(v => v.subject)


    // Rebuild series: each workload becomes a series
    const newSeries = workloads.map(function(workload) {
        const data = orderSubject.map(function(subject) {
            return dataMap[workload][subject];
        });

        return {
            name: workload,
            type: 'bar',
            data: data,
            label: {
                show: true,
                position: 'top',
                formatter: '{c}',
                fontSize: 10
            }
        };
    });

    // Clear and rebuild chart with new structure
    chartInstance.clear();
    const newOption = JSON.parse(JSON.stringify(originalData));
    newOption.xAxis[0].data = orderSubject;
    newOption.series = newSeries;

    chartInstance.setOption(newOption);
}

// Apply sort to all charts
function sortAllCharts(sortOrder) {
    const chartElements = document.querySelectorAll('.item');

    chartElements.forEach(function(chartDom) {
        const instance = echarts.getInstanceByDom(chartDom);

        if (instance) {
            sortChart(instance, sortOrder);
        } else {
            console.warn('No ECharts instance found for:', chartDom.id);
        }
    });

    // Update button states
    document.querySelectorAll('.sort-btn').forEach(function(btn) {
        btn.classList.remove('active');
    });
    const activeBtn = document.querySelector('.sort-btn[data-sort="' + sortOrder + '"]');
    if (activeBtn) {
        activeBtn.classList.add('active');
    }
}

// Add resize event handler to make charts responsive
window.addEventListener('resize', function() {
    // Resize all charts when window size changes
    const charts = document.querySelectorAll('.item');
    charts.forEach(function(chart) {
        const instance = echarts.getInstanceByDom(chart);
        if (instance) {
            instance.resize();
        }
    });
});

// Initialize sidebar functionality when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    const sidebar = document.getElementById('bench-sidebar');
    const groups = document.querySelectorAll('.bench-indicator');
    const chartSections = document.querySelectorAll('.chart-section');
    const minimizeBtn = document.getElementById('minimize-btn');

    // Only show sidebar if there are multiple groups
    if (sidebar && groups.length <= 1) {
        sidebar.style.display = 'none';
    }

    // Make sidebar draggable
    let isDragging = false;
    let offsetX, offsetY;

    function startDrag(e) {
        // Only start drag if it's the sidebar header or the sidebar itself when minimized
        if (e.target.closest('.sidebar-header') ||
            (sidebar.classList.contains('minimized') && e.target.closest('.sidebar'))) {
            isDragging = true;

            // Get the current position of the sidebar
            const sidebarRect = sidebar.getBoundingClientRect();

            // Calculate the offset from the mouse position to the sidebar position
            offsetX = e.clientX - sidebarRect.left;
            offsetY = e.clientY - sidebarRect.top;

            // Prevent text selection during drag
            e.preventDefault();
        }
    }

    function drag(e) {
        if (isDragging) {
            // Calculate new position
            const x = e.clientX - offsetX;
            const y = e.clientY - offsetY;

            // Apply new position
            sidebar.style.left = x + 'px';
            sidebar.style.top = y + 'px';
            sidebar.style.right = 'auto';
            sidebar.style.transform = 'none';
        }
    }

    function stopDrag() {
        isDragging = false;
    }

    // Add event listeners for dragging
    sidebar.addEventListener('mousedown', startDrag);
    document.addEventListener('mousemove', drag);
    document.addEventListener('mouseup', stopDrag);

    // Minimize/maximize sidebar
    minimizeBtn.addEventListener('click', function() {
        sidebar.classList.toggle('minimized');
        this.textContent = sidebar.classList.contains('minimized') ? '◀' : '▶';
    });

    // Add click event to each indicator
    groups.forEach(function(indicator) {
        indicator.addEventListener('click', function() {
            const targetId = this.getAttribute('data-target');
            const targetElement = document.getElementById(targetId);

            if (targetElement) {
                // Scroll to the target element
                targetElement.scrollIntoView({ behavior: 'smooth' });

                // Update active state
                groups.forEach(ind => ind.classList.remove('active'));
                this.classList.add('active');
            }
        });
    });

    // Update active indicator on scroll
    window.addEventListener('scroll', function() {
        let currentSection = '';

        chartSections.forEach(function(section) {
            const sectionTop = section.offsetTop;

            if (pageYOffset >= (sectionTop - 100)) {
                currentSection = section.getAttribute('id');
            }
        });

        groups.forEach(function(indicator) {
            indicator.classList.remove('active');
            if (indicator.getAttribute('data-target') === currentSection) {
                indicator.classList.add('active');
            }
        });
    });
});
