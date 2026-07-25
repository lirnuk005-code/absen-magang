document.addEventListener('DOMContentLoaded', () => {
  // DOM Elements
  const loginScreen = document.getElementById('login-screen');
  const dashboardScreen = document.getElementById('dashboard-screen');
  const loginForm = document.getElementById('login-form');
  const userDisplayName = document.getElementById('user-display-name');
  const btnLogout = document.getElementById('btn-logout');

  // Status Displays
  const liveClock = document.getElementById('live-clock');
  const timeStatusBadge = document.getElementById('time-status-badge');
  const geoDistance = document.getElementById('geo-distance');
  const geoStatusBadge = document.getElementById('geo-status-badge');
  const geoCoords = document.getElementById('geo-coords');
  const currentIpDisplay = document.getElementById('current-ip-display');
  const ipStatusBadge = document.getElementById('ip-status-badge');
  const regIpText = document.getElementById('reg-ip-text');
  const ipRegisterBanner = document.getElementById('ip-register-banner');
  const btnRegisterIp = document.getElementById('btn-register-ip');

  // Absen Action
  const btnAbsen = document.getElementById('btn-absen');
  const mockLocationToggle = document.getElementById('mock-location-toggle');
  const resultAlert = document.getElementById('result-alert');
  const logsTableBody = document.getElementById('logs-table-body');
  const btnRefreshLogs = document.getElementById('btn-refresh-logs');

  let currentUser = null;
  let currentLat = 0;
  let currentLng = 0;

  // Target Location: Jalan Gatot Subroto I, Denpasar (-8.6366, 115.2223)
  const TARGET_LAT = -8.6366;
  const TARGET_LNG = 115.2223;

  // 1. Live Clock Function (WITA UTC+8)
  function startClock() {
    setInterval(() => {
      const now = new Date();
      // Format WITA time
      const options = { timeZone: 'Asia/Makassar', hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' };
      const timeStr = new Intl.DateTimeFormat('id-ID', options).format(now);
      liveClock.textContent = timeStr;

      // Check Cutoff (08:30 WITA)
      const parts = timeStr.split('.');
      const hour = parseInt(parts[0], 10);
      const minute = parseInt(parts[1], 10);

      if (hour < 8 || (hour === 8 && minute <= 30)) {
        timeStatusBadge.className = 'badge badge-success';
        timeStatusBadge.textContent = 'Absen Dibuka';
      } else {
        timeStatusBadge.className = 'badge badge-danger';
        timeStatusBadge.textContent = 'Absen Ditutup (>08:30)';
      }
    }, 1000);
  }

  // 2. Haversine Distance Helper
  function calculateHaversine(lat1, lon1, lat2, lon2) {
    const R = 6371000; // meters
    const dLat = (lat2 - lat1) * Math.PI / 180;
    const dLon = (lon2 - lon1) * Math.PI / 180;
    const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
              Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
              Math.sin(dLon / 2) * Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  }

  // 3. Geolocation Handler
  function updateLocation() {
    if (mockLocationToggle.checked) {
      currentLat = TARGET_LAT;
      currentLng = TARGET_LNG;
      geoDistance.textContent = '0.0 m';
      geoCoords.textContent = `Lat: ${TARGET_LAT.toFixed(4)} | Lng: ${TARGET_LNG.toFixed(4)} (Simulasi Denpasar)`;
      geoStatusBadge.className = 'badge badge-success';
      geoStatusBadge.textContent = 'Dalam Area Jln Gatsu I';
      return;
    }

    if ('geolocation' in navigator) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          currentLat = position.coords.latitude;
          currentLng = position.coords.longitude;
          const dist = calculateHaversine(currentLat, currentLng, TARGET_LAT, TARGET_LNG);
          geoDistance.textContent = `${dist.toFixed(1)} m`;
          geoCoords.textContent = `Lat: ${currentLat.toFixed(4)} | Lng: ${currentLng.toFixed(4)}`;

          if (dist <= 100) {
            geoStatusBadge.className = 'badge badge-success';
            geoStatusBadge.textContent = 'Dalam Radius (<=100m)';
          } else {
            geoStatusBadge.className = 'badge badge-danger';
            geoStatusBadge.textContent = 'Di Luar Radius (>100m)';
          }
        },
        (error) => {
          geoStatusBadge.className = 'badge badge-warning';
          geoStatusBadge.textContent = 'GPS Diperlukan / Simulasi';
          geoCoords.textContent = 'Gunakan toggle "Simulasi Lokasi" jika GPS browser diblokir';
        },
        { enableHighAccuracy: true }
      );
    }
  }

  mockLocationToggle.addEventListener('change', updateLocation);

  // 4. Fetch User Data (/api/me)
  async function fetchUserData() {
    try {
      const res = await fetch('/api/me');
      const json = await res.json();

      if (json.success) {
        currentUser = json.data;
        userDisplayName.textContent = currentUser.username;
        currentIpDisplay.textContent = currentUser.current_ip;

        if (currentUser.registered_ip) {
          regIpText.textContent = currentUser.registered_ip;
          ipStatusBadge.className = 'badge badge-success';
          ipStatusBadge.textContent = '1 IP Terdaftar (Terkunci)';
          ipRegisterBanner.style.display = 'none'; // Hide register button if already has 1 IP
        } else {
          regIpText.textContent = 'Belum Terdaftar (Max 1 IP)';
          ipStatusBadge.className = 'badge badge-danger';
          ipStatusBadge.textContent = 'Belum Terdaftar';
          ipRegisterBanner.style.display = 'flex';
        }

        loginScreen.classList.remove('active');
        dashboardScreen.classList.add('active');
        updateLocation();
        loadLogs();
      } else {
        loginScreen.classList.add('active');
        dashboardScreen.classList.remove('active');
      }
    } catch (err) {
      console.error('Error fetching me:', err);
    }
  }

  // 5. Login Form Handler
  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;

    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });
      const json = await res.json();

      if (json.success) {
        fetchUserData();
      } else {
        alert(json.message || 'Login gagal!');
      }
    } catch (err) {
      alert('Gagal terhubung ke server!');
    }
  });

  // 6. Register IP Button Handler
  btnRegisterIp.addEventListener('click', async () => {
    if (!confirm('Daftarkan IP Anda saat ini secara permanen ke akun ini?')) return;

    try {
      const res = await fetch('/api/register-ip', { method: 'POST' });
      const json = await res.json();

      if (json.success) {
        alert(json.message);
        fetchUserData();
      } else {
        alert('Gagal registrasi IP: ' + json.message);
      }
    } catch (err) {
      alert('Terjadi kesalahan jaringan!');
    }
  });

  // 7. Absen Button Handler
  btnAbsen.addEventListener('click', async () => {
    resultAlert.className = 'alert-box hidden';
    btnAbsen.disabled = true;

    // Use current location or target location if mock
    const lat = currentLat || TARGET_LAT;
    const lng = currentLng || TARGET_LNG;

    try {
      const res = await fetch('/api/absen', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ latitude: lat, longitude: lng })
      });
      const json = await res.json();

      resultAlert.classList.remove('hidden');
      if (json.success) {
        resultAlert.className = 'alert-box alert-success';
        resultAlert.innerHTML = `<i class="fa-solid fa-circle-check"></i> ${json.message}`;
      } else {
        resultAlert.className = 'alert-box alert-error';
        resultAlert.innerHTML = `<i class="fa-solid fa-triangle-exclamation"></i> ${json.message}`;
      }
      loadLogs();
    } catch (err) {
      resultAlert.classList.remove('hidden');
      resultAlert.className = 'alert-box alert-error';
      resultAlert.innerHTML = `<i class="fa-solid fa-triangle-exclamation"></i> Gagal terhubung ke server!`;
    } finally {
      btnAbsen.disabled = false;
    }
  });

  // 8. Fetch Attendance Logs Table
  async function loadLogs() {
    try {
      const res = await fetch('/api/logs');
      const json = await res.json();

      if (json.success && json.data) {
        if (json.data.length === 0) {
          logsTableBody.innerHTML = `<tr><td colspan="6" class="text-center">Belum ada data presensi hari ini.</td></tr>`;
          return;
        }

        logsTableBody.innerHTML = json.data.map(log => {
          const dateObj = new Date(log.check_in_time);
          const timeFormatted = dateObj.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });

          let badgeClass = 'badge-success';
          if (log.status !== 'SUCCESS') badgeClass = 'badge-danger';

          return `
            <tr>
              <td><b>${timeFormatted}</b> WITA</td>
              <td>${log.username}</td>
              <td><code>${log.ip_address}</code></td>
              <td>${log.distance_meters.toFixed(1)} m</td>
              <td><span class="badge ${badgeClass}">${log.status}</span></td>
              <td>${log.notes}</td>
            </tr>
          `;
        }).join('');
      }
    } catch (err) {
      console.error('Error loading logs:', err);
    }
  }

  btnRefreshLogs.addEventListener('click', loadLogs);

  // 9. Logout
  btnLogout.addEventListener('click', async () => {
    await fetch('/api/logout', { method: 'POST' });
    loginScreen.classList.add('active');
    dashboardScreen.classList.remove('active');
  });

  // Initialize
  startClock();
  fetchUserData();
  setInterval(updateLocation, 5000);
});
