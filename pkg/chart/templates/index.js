// Store original chart data for each chart instance
const originalChartData = new Map();

// Chart state management - Initialize with flag settings if available
const chartSettingState = {
  sortOrder: typeof flagSettings !== 'undefined' ? flagSettings.sortOrder : 'default',
  showLabels: typeof flagSettings !== 'undefined' ? flagSettings.showLabels : false,
};

// Sort chart intelligently based on structure
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
  const subjects = originalData.series.map(function (s) { return s.name; });

  // Determine the structure:
  // - If X-axis has meaningful data (non-empty, multiple items), workloads are on X-axis
  // - If X-axis is empty or single empty item, subjects are in series
  const hasRealWorkloads = workloads && workloads.length > 1 && workloads[0] !== '';

  if (hasRealWorkloads) {
    const dataMap = {};
    workloads.forEach(function (workload, wIdx) {
      dataMap[workload] = {};
      originalData.series.forEach(function (series) {
        const dataPoint = series.data[wIdx];
        const value = typeof dataPoint === 'object' ? parseFloat(dataPoint.value) : parseFloat(dataPoint);
        dataMap[workload][series.name] = isNaN(value) ? 0 : value;
      });
    });

    // For each workload, sort subjects by their values
    const subjectMap = new Map()

    workloads.forEach(function (workload) {
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
    const newSeries = workloads.map(function (workload) {
      const data = orderSubject.map(function (subject) {
        return dataMap[workload][subject];
      });

      return {
        name: workload,
        type: 'bar',
        data: data,
        label: {
          show: chartSettingState.showLabels,
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

    // Ensure proper label rendering for many bars
    if (!newOption.xAxis[0].axisLabel) {
      newOption.xAxis[0].axisLabel = {};
    }
    newOption.xAxis[0].axisLabel.rotate = 0;
    newOption.xAxis[0].axisLabel.interval = 0;

    chartInstance.setOption(newOption);
  } else {
    // Calculate total value for each series (subject)
    const seriesTotals = originalData.series.map(function (series, idx) {
      let total = 0;
      series.data.forEach(function (dataPoint) {
        const value = typeof dataPoint === 'object' ? parseFloat(dataPoint.value) : parseFloat(dataPoint);
        total += isNaN(value) ? 0 : value;
      });
      return {
        series: series,
        name: series.name,
        total: total,
        originalIndex: idx
      };
    });


    // Sort series by total value
    seriesTotals.sort(function (a, b) {
      if (sortOrder === 'asc') {
        return a.total - b.total;
      }
      return b.total - a.total;
    });


    // Rebuild series in sorted order, preserving all properties including colors
    const sortedSeries = seriesTotals.map(function (item) {
      const series = JSON.parse(JSON.stringify(item.series));
      // Ensure label state is current
      if (series.label) {
        series.label.show = chartSettingState.showLabels;
      }
      return series;
    });

    // Update chart (NO SWAP - just reorder series)
    const newOption = JSON.parse(JSON.stringify(originalData));
    newOption.series = sortedSeries;

    chartInstance.clear();
    chartInstance.setOption(newOption);
  }
}

// Apply sort to all charts
function sortAllCharts(sortOrder) {
  const chartElements = document.querySelectorAll('.item');

  chartElements.forEach(function (chartDom) {
    const instance = echarts.getInstanceByDom(chartDom);

    if (instance) {
      sortChart(instance, sortOrder);
    } else {
      console.warn('No ECharts instance found for:', chartDom.id);
    }
  });

  // Update button states
  document.querySelectorAll('.sort-btn').forEach(function (btn) {
    btn.classList.remove('active');
  });
  const activeBtn = document.querySelector('.sort-btn[data-sort="' + sortOrder + '"]');
  if (activeBtn) {
    activeBtn.classList.add('active');
  }
}


// Toggle chart labels
function toggleLabels() {
  chartSettingState.showLabels = !chartSettingState.showLabels;

  // Update toggle button state
  const toggleBtn = document.querySelector('.toggle-switch');
  if (toggleBtn) {
    if (chartSettingState.showLabels) {
      toggleBtn.classList.add('active');
    } else {
      toggleBtn.classList.remove('active');
    }
  }

  // Update all charts
  const chartElements = document.querySelectorAll('.item');
  chartElements.forEach(function (chartDom) {
    const instance = echarts.getInstanceByDom(chartDom);
    if (instance) {
      const option = instance.getOption();

      // Update label visibility for all series
      if (option.series) {
        option.series.forEach(function (series) {
          if (!series.label) {
            series.label = {};
          }
          series.label.show = chartSettingState.showLabels;
        });

        instance.setOption(option);
      }
    }
  });
}

// Toggle control panel visibility
function toggleControlPanel() {
  const panel = document.getElementById('controlPanel');
  const toggleBtn = document.querySelector('.control-toggle');

  if (panel) {
    const isOpen = panel.classList.contains('open');

    if (isOpen) {
      panel.classList.remove('open');
      if (toggleBtn) {
        toggleBtn.style.opacity = '1';
      }
    } else {
      panel.classList.add('open');
      if (toggleBtn) {
        toggleBtn.style.opacity = '0';
      }
    }
  }
}

// Toggle bench groups panel visibility
function toggleBenchGroups() {
  const panel = document.getElementById('bench-groups');
  const toggleBtn = document.querySelector('.groups-toggle');

  if (panel && toggleBtn) {
    const isOpen = panel.classList.contains('open');

    if (isOpen) {
      panel.classList.remove('open');
      toggleBtn.classList.remove('hidden');
    } else {
      panel.classList.add('open');
      toggleBtn.classList.add('hidden');
    }
  }
}

// Close panels when clicking outside
document.addEventListener('click', function (e) {
  // Handle control panel
  const controlPanel = document.getElementById('controlPanel');
  const controlToggleBtn = document.querySelector('.control-toggle');

  if (controlPanel && controlPanel.classList.contains('open')) {
    if (!controlPanel.contains(e.target) && !controlToggleBtn.contains(e.target)) {
      controlPanel.classList.remove('open');
      if (controlToggleBtn) {
        controlToggleBtn.style.opacity = '1';
      }
    }
  }

  // Handle bench groups panel
  const benchGroupsPanel = document.getElementById('bench-groups');
  const groupsToggleBtn = document.querySelector('.groups-toggle');

  if (benchGroupsPanel && benchGroupsPanel.classList.contains('open')) {
    if (!benchGroupsPanel.contains(e.target) && !groupsToggleBtn?.contains(e.target)) {
      benchGroupsPanel.classList.remove('open');
      if (groupsToggleBtn) {
        groupsToggleBtn.classList.remove('hidden');
      }
    }
  }
});

// Add resize event handler to make charts responsive
let resizeTimeout;
window.addEventListener('resize', function () {
  // Debounce resize for better performance
  clearTimeout(resizeTimeout);
  resizeTimeout = setTimeout(function () {
    const charts = document.querySelectorAll('.item');
    charts.forEach(function (chart) {
      const instance = echarts.getInstanceByDom(chart);
      if (instance) {
        instance.resize();
      }
    });
  }, 150);
});

// Initialize bench groups functionality when DOM is loaded
document.addEventListener('DOMContentLoaded', function () {
  const benchGroups = document.getElementById('bench-groups');
  const groupItems = document.querySelectorAll('.group-item');
  const chartSections = document.querySelectorAll('.chart-section');
  const groupsToggleBtn = document.querySelector('.groups-toggle');

  if (!benchGroups) return;

  // Add click event to each group item
  groupItems.forEach(function (item) {
    item.addEventListener('click', function () {
      const targetId = this.getAttribute('data-target');
      const targetElement = document.getElementById(targetId);

      if (targetElement) {
        // Scroll to the target element
        targetElement.scrollIntoView({ behavior: 'smooth', block: 'start' });

        // Update active state
        groupItems.forEach(i => i.classList.remove('active'));
        this.classList.add('active');

        // Close the panel after navigation
        benchGroups.classList.remove('open');
        if (groupsToggleBtn) {
          groupsToggleBtn.classList.remove('hidden');
        }
      }
    });
  });

  // Update active group item on scroll with debouncing
  let scrollTimeout;
  window.addEventListener('scroll', function () {
    if (scrollTimeout) {
      window.cancelAnimationFrame(scrollTimeout);
    }

    scrollTimeout = window.requestAnimationFrame(function () {
      let currentSection = '';

      chartSections.forEach(function (section) {
        const sectionTop = section.offsetTop;

        if (window.pageYOffset >= (sectionTop - 100)) {
          currentSection = section.getAttribute('id');
        }
      });

      groupItems.forEach(function (item) {
        item.classList.remove('active');
        if (item.getAttribute('data-target') === currentSection) {
          item.classList.add('active');
        }
      });
    });
  });
});
