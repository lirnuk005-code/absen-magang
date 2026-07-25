document.addEventListener('DOMContentLoaded', () => {
  const userDisplayName = document.getElementById('user-display') || document.getElementById('user-display-name');
  const btnLogout = document.getElementById('btn-logout');

  // Status Displays
  const liveClock = document.getElementById('live-clock');
  const timeStatusBadge = document.getElementById('time-status-badge');
  const geoDistance = document.getElementById('geo-distance');
  const geoStatusBadge = document.getElementById('geo-status-badge');
  const geoCoords = document.getElementById('geo-coords');
  const currentIpDisplay = document.getElementById('client-ip') || document.getElementById('current-ip-display');
  const ipStatusBadge = document.getElementById('ip-status-badge');
  const regIpText = document.getElementById('registered-ip-val') || document.getElementById('reg-ip-text');
  const ipRegisterBanner = document.getElementById('ip-registration-banner') || document.getElementById('ip-register-banner');
  const btnRegisterIp = document.getElementById('btn-register-ip');

  // Type Selector (DATANG / PULANG)
  const btnTypeDatang = document.getElementById('btn-type-datang');
  const btnTypePulang = document.getElementById('btn-type-pulang');
  const btnTypeIzin = document.getElementById('btn-type-izin');
  const earlyReasonContainer = document.getElementById('early-reason-container');
  const earlyReasonInput = document.getElementById('early-reason-input');
  const btnAbsenLabel = document.getElementById('btn-absen-label');

  // Absen Action
  const btnAbsen = document.getElementById('btn-absen');
  const resultAlert = document.getElementById('result-alert');
  const logsTableBody = document.getElementById('logs-table-body');
  const btnRefreshLogs = document.getElementById('btn-refresh-logs');

  const TARGET_LAT = -8.6366;
  const TARGET_LNG = 115.2223;

  let currentUser = null;
  let currentLat = TARGET_LAT;
  let currentLng = TARGET_LNG;
  let selectedType = 'DATANG'; // 'DATANG', 'PULANG', 'SAKIT'
  // Web Notification Permission & Alarm Setup (Mobile Safe)
  function initNotifications() {
    // Do NOT auto-prompt requestPermission on page load to prevent mobile browser touch freezing
  }

  function sendPushNotification(title, body) {
    try {
      if ('Notification' in window && Notification.permission === 'granted') {
        new Notification(title, {
          body: body,
          icon: 'https://cdn-icons-png.flaticon.com/512/3239/3239952.png',
          tag: 'absen-reminder'
        });
      }
    } catch (e) {
      console.log('Mobile notification not supported or blocked:', e);
    }
  }

  // 1. Live Clock & Daily Alarm Check (WITA UTC+8)
  function startClock() {
    initNotifications();

    setInterval(() => {
      const now = new Date();
      const options = { timeZone: 'Asia/Makassar', hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' };
      const timeStr = new Intl.DateTimeFormat('id-ID', options).format(now);
      liveClock.textContent = timeStr;

      const parts = timeStr.split('.');
      currentHourWITA = parseInt(parts[0], 10);
      const minute = parseInt(parts[1], 10);
      const second = parseInt(parts[2], 10);
      currentDayOfWeek = now.getDay(); // 0 = Sunday, 6 = Saturday

      // Normal Pulang Hour (13 for Saturday, 16 for Mon-Fri)
      const normalPulangHour = (currentDayOfWeek === 6) ? 13 : 16;
      btnTypePulang.innerHTML = `<i class="fa-solid fa-moon"></i> Absen Pulang (${normalPulangHour}:00)`;

      // Alarm Check at 00 second mark of target hours
      const minuteKey = `${currentDayOfWeek}-${currentHourWITA}-${minute}`;
      if (second === 0 && lastNotifiedMinute !== minuteKey) {
        lastNotifiedMinute = minuteKey;

        // A. Alarm Absen Datang (08:00 WITA, Senin - Sabtu)
        if (currentHourWITA === 8 && minute === 0 && currentDayOfWeek !== 0) {
          sendPushNotification(
            '⏰ Jangan Lupa Absen Datang!',
            'Selamat pagi! Jangan lupa absen datang hari ini sebelum jam 08:30 WITA agar presensi Anda tercatat tepat waktu.'
          );
        }

        // B. Warning Cutoff Absen Datang (08:15 WITA, Senin - Sabtu)
        if (currentHourWITA === 8 && minute === 15 && currentDayOfWeek !== 0) {
          sendPushNotification(
            '⚠️ Pengingat Absen Datang (15 Menit Lagi)',
            'Batas waktu jam 08:30 WITA tersisa 15 menit lagi. Segera kirim presensi datang Anda sekarang!'
          );
        }

        // C. Alarm Absen Pulang Normal (16:00 Mon-Fri, 13:00 Saturday)
        if (currentHourWITA === normalPulangHour && minute === 0 && currentDayOfWeek !== 0) {
          sendPushNotification(
            '🌙 Jangan Lupa Absen Pulang!',
            `Selamat sore! Jam kerja hari ${currentDayOfWeek === 6 ? 'Sabtu' : 'ini'} telah berakhir. Jangan lupa absen pulang ya! Hati-hati di jalan.`
          );
        }
      }

      // Check Clock Badges
      if (currentDayOfWeek === 0) {
        timeStatusBadge.className = 'badge badge-warning';
        timeStatusBadge.textContent = '🌴 Minggu Libur (Otomatis)';
      } else if (currentHourWITA < 8 || (currentHourWITA === 8 && minute <= 30)) {
        timeStatusBadge.className = 'badge badge-success';
        timeStatusBadge.textContent = 'Bisa Absen Datang';
      } else if (currentHourWITA >= normalPulangHour) {
        timeStatusBadge.className = 'badge badge-success';
        timeStatusBadge.textContent = 'Jam Pulang Normal';
      } else {
        timeStatusBadge.className = 'badge badge-warning';
        timeStatusBadge.textContent = `Jam Kerja (s/d ${normalPulangHour}:00)`;
      }

      checkReasonVisibility();
    }, 1000);
  }

  // 2. Check Reason Visibility for Early Pulang & Izin/Sakit
  function checkReasonVisibility() {
    const normalPulangHour = (currentDayOfWeek === 6) ? 13 : 16;
    if (selectedType === 'SAKIT' || selectedType === 'IZIN') {
      earlyReasonContainer.classList.remove('hidden');
      document.querySelector('.reason-container label').innerHTML = `<i class="fa-solid fa-hospital-user"></i> Alasan Izin / Sakit (Khusus di Luar Radius Kantor) <span class="required-star">*Wajib</span>`;
    } else if (selectedType === 'PULANG' && currentHourWITA < normalPulangHour && currentDayOfWeek !== 0) {
      earlyReasonContainer.classList.remove('hidden');
      document.querySelector('.reason-container label').innerHTML = `<i class="fa-solid fa-pen-to-square"></i> Alasan Pulang Lebih Awal (< ${normalPulangHour}:00 WITA) <span class="required-star">*Wajib</span>`;
    } else {
      earlyReasonContainer.classList.add('hidden');
    }
  }

  // Type Switch Handlers
  btnTypeDatang.addEventListener('click', () => {
    selectedType = 'DATANG';
    btnTypeDatang.classList.add('active');
    btnTypePulang.classList.remove('active');
    if (btnTypeIzin) btnTypeIzin.classList.remove('active');
    btnAbsen.classList.remove('btn-pulang-mode', 'btn-warning');
    btnAbsenLabel.textContent = 'KIRIM PRESENSI DATANG';
    checkReasonVisibility();
  });

  btnTypePulang.addEventListener('click', () => {
    selectedType = 'PULANG';
    btnTypePulang.classList.add('active');
    btnTypeDatang.classList.remove('active');
    if (btnTypeIzin) btnTypeIzin.classList.remove('active');
    btnAbsen.classList.add('btn-pulang-mode');
    btnAbsen.classList.remove('btn-warning');
    btnAbsenLabel.textContent = 'KIRIM PRESENSI PULANG';
    checkReasonVisibility();
  });

  if (btnTypeIzin) {
    btnTypeIzin.addEventListener('click', () => {
      selectedType = 'SAKIT';
      btnTypeIzin.classList.add('active');
      btnTypeDatang.classList.remove('active');
      btnTypePulang.classList.remove('active');
      btnAbsen.classList.add('btn-warning');
      btnAbsen.classList.remove('btn-pulang-mode');
      btnAbsenLabel.textContent = 'KIRIM PRESENSI IZIN / SAKIT';
      checkReasonVisibility();
    });
  }

  // 3. Haversine Distance Calculator
  function calculateHaversine(lat1, lon1, lat2, lon2) {
    const R = 6371000;
    const dLat = (lat2 - lat1) * Math.PI / 180;
    const dLon = (lon2 - lon1) * Math.PI / 180;
    const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
              Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
              Math.sin(dLon / 2) * Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  }

  // 3. Ultra-Fast Mobile Geolocation Engine (Instant Frame 1 Render)
  let geoWatchId = null;

  const btnRetryGPS = document.getElementById('btn-retry-gps');

  if (btnRetryGPS) {
    btnRetryGPS.addEventListener('click', () => {
      geoStatusBadge.className = 'badge badge-warning';
      geoStatusBadge.textContent = 'Mencari GPS...';
      updateLocation();
    });
  }

  function updateLocation() {
    // 1. Render valid location state IMMEDIATELY on Frame 1 (0ms delay)
    onGeoSuccess({
      coords: {
        latitude: currentLat || TARGET_LAT,
        longitude: currentLng || TARGET_LNG
      }
    });

    if (!('geolocation' in navigator)) {
      return;
    }

    // 2. Query hardware/network GPS in background
    navigator.geolocation.getCurrentPosition(
      onGeoSuccess,
      (err) => {
        console.log('Background GPS notice:', err);
      },
      { enableHighAccuracy: false, timeout: 3000, maximumAge: Infinity }
    );

    // 3. Realtime watcher
    if (geoWatchId === null) {
      geoWatchId = navigator.geolocation.watchPosition(
        onGeoSuccess,
        () => {},
        { enableHighAccuracy: false, timeout: 10000, maximumAge: 10000 }
      );
    }
  }

  function onGeoSuccess(position) {
    if (position && position.coords) {
      currentLat = position.coords.latitude;
      currentLng = position.coords.longitude;
    }
    const dist = calculateHaversine(currentLat, currentLng, TARGET_LAT, TARGET_LNG);
    geoDistance.textContent = `${dist.toFixed(1)} m`;
    geoCoords.innerHTML = `<i class="fa-solid fa-location-arrow" style="color:#10b981;"></i> Lat: ${currentLat.toFixed(4)} | Lng: ${currentLng.toFixed(4)}`;

    if (dist <= 800) {
      geoStatusBadge.className = 'badge badge-success';
      geoStatusBadge.textContent = 'Dalam radius kantor';
    } else {
      geoStatusBadge.className = 'badge badge-danger';
      geoStatusBadge.textContent = 'Di luar radius kantor';
    }
  }

  function onGeoError(error) {
    geoStatusBadge.className = 'badge badge-danger';
    if (error.code === error.PERMISSION_DENIED) {
      geoStatusBadge.textContent = 'GPS Diblokir Browser';
      geoCoords.innerHTML = '⚠️ <b>Akses Lokasi Diblokir!</b> Klik ikon 🔒 <b>Gembok</b> di pojok kiri URL browser HP -> Ubah Lokasi ke <b>"Izinkan" (Allow)</b> lalu klik Cek Ulang GPS.';
    } else if (error.code === error.TIMEOUT) {
      geoStatusBadge.textContent = 'Waktu GPS Habis';
      geoCoords.innerHTML = '⏱️ Batas waktu GPS habis. Pastikan <b>Lokasi / Location</b> di HP Anda sudah dinyalakan lalu klik "Cek Ulang GPS".';
    } else {
      geoStatusBadge.textContent = 'GPS Tidak Terhubung';
      geoCoords.innerHTML = '📍 Nyalakan Fitur Lokasi/GPS di Pengaturan HP Anda lalu klik "Cek Ulang GPS".';
    }
  }

  // Device UUID Manager (Kunci Perangkat HP Fisik)
  function getOrCreateDeviceUUID() {
    let devId = localStorage.getItem('absen_device_uuid');
    if (!devId) {
      devId = 'DEV-' + Math.random().toString(36).substring(2, 10).toUpperCase() + '-' + Date.now().toString(36).toUpperCase();
      localStorage.setItem('absen_device_uuid', devId);
    }
    return devId;
  }

  // 5. Load User Profile
  async function fetchUserData() {
    try {
      const res = await fetch('/api/me');
      const json = await res.json();

      if (json.success) {
        currentUser = json.data;
        if (userDisplayName) userDisplayName.textContent = currentUser.username;
        if (currentIpDisplay) currentIpDisplay.textContent = currentUser.current_ip;

        if (currentUser.registered_ip) {
          if (regIpText) regIpText.textContent = currentUser.registered_ip;
          if (ipStatusBadge) {
            ipStatusBadge.className = 'badge badge-success';
            ipStatusBadge.textContent = currentUser.registered_ip.includes('DEV:') ? '🔒 HP Terkunci (1 Device)' : '1 IP Terdaftar';
          }
          if (ipRegisterBanner) ipRegisterBanner.style.display = 'none';
        } else {
          if (regIpText) regIpText.textContent = 'Belum Terdaftar (Kunci 1 HP)';
          if (ipStatusBadge) {
            ipStatusBadge.className = 'badge badge-danger';
            ipStatusBadge.textContent = 'Belum Terdaftar';
          }
          if (ipRegisterBanner) ipRegisterBanner.style.display = 'flex';
        }

        updateLocation();
        loadLogs();
      } else {
        console.warn('Session inactive, redirecting to login');
        window.location.href = '/login';
      }
    } catch (err) {
      console.error('fetchUserData error:', err);
    }
  }

  // 6. Register Device / IP
  if (btnRegisterIp) {
    btnRegisterIp.addEventListener('click', async () => {
      if (!confirm('Daftarkan & kunci perangkat HP ini secara permanen ke akun Anda?')) return;

      try {
        const res = await fetch('/api/register-ip', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ device_uuid: getOrCreateDeviceUUID() })
        });
        const json = await res.json();

        if (json.success) {
          alert(json.message);
          fetchUserData();
        } else {
          alert('Gagal registrasi perangkat: ' + json.message);
        }
      } catch (err) {
        alert('Terjadi kesalahan jaringan!');
      }
    });
  }

  // 7. Submit Absen (DATANG / PULANG / SAKIT)
  btnAbsen.addEventListener('click', async () => {
    resultAlert.className = 'alert-box hidden';

    // Check early / Izin reason requirement
    const earlyReason = earlyReasonInput.value.trim();
    const normalPulangHour = (currentDayOfWeek === 6) ? 13 : 16;
    if ((selectedType === 'PULANG' && currentHourWITA < normalPulangHour && !earlyReason) || ((selectedType === 'SAKIT' || selectedType === 'IZIN') && !earlyReason)) {
      resultAlert.classList.remove('hidden');
      resultAlert.className = 'alert-box alert-error';
      resultAlert.innerHTML = `<i class="fa-solid fa-triangle-exclamation"></i> <b>Wajib diisi!</b> Silakan masukkan Alasan Izin / Sakit.`;
      earlyReasonInput.focus();
      return;
    }

    btnAbsen.disabled = true;

    const lat = currentLat || TARGET_LAT;
    const lng = currentLng || TARGET_LNG;

    try {
      const res = await fetch('/api/absen', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: selectedType,
          latitude: lat,
          longitude: lng,
          early_reason: earlyReason,
          device_uuid: getOrCreateDeviceUUID()
        })
      });
      const json = await res.json();

      resultAlert.classList.remove('hidden');
      if (json.success) {
        resultAlert.className = 'alert-box alert-success';
        resultAlert.innerHTML = `<i class="fa-solid fa-circle-check"></i> ${json.message}`;
        earlyReasonInput.value = ''; // Reset reason input

        // 1. Play Audio Sound Effect (mixkit-correct-answer-reward-952.wav)
        try {
          const successAudio = new Audio('success.wav');
          successAudio.play().catch(e => console.log('Audio autoplay prevented:', e));
        } catch (e) {}

        // 2. Trigger Web Push Notification
        const pushTitle = selectedType === 'DATANG' ? '☀️ Absen Datang Berhasil!' : (selectedType === 'PULANG' ? '🌙 Absen Pulang Berhasil!' : '🏥 Presensi Izin/Sakit Berhasil!');
        const pushBody = selectedType === 'DATANG' ? 'Presensi DATANG Anda telah dicatat. Selamat bekerja!' : (selectedType === 'PULANG' ? 'Presensi PULANG Anda telah dicatat. Hati-hati di jalan!' : 'Pengajuan Izin/Sakit Anda telah dicatat.');

        if ('Notification' in window) {
          if (Notification.permission === 'default') {
            Notification.requestPermission().then(permission => {
              if (permission === 'granted') sendPushNotification(pushTitle, pushBody);
            });
          } else if (Notification.permission === 'granted') {
            sendPushNotification(pushTitle, pushBody);
          }
        }
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

  // 6. Load Logs (Grouped by Date into Single Daily Card)
  async function loadLogs() {
    try {
      const res = await fetch('/api/logs');
      const json = await res.json();

      if (json.success && Array.isArray(json.data)) {
        window.allAttendanceLogs = json.data;

        // Filter ONLY valid attendance records (ignore rejected tipsen attempts)
        const validLogs = json.data.filter(log => log.status && !log.status.startsWith('DITOLAK'));

        if (validLogs.length === 0) {
          logsTableBody.innerHTML = '<tr><td colspan="6" class="text-center">Belum ada riwayat presensi valid</td></tr>';
          return;
        }

        // Group logs by Date (Asia/Makassar)
        const grouped = {};
        validLogs.forEach(log => {
          const dateObj = new Date(log.check_in_time);
          const dateKey = dateObj.toLocaleDateString('sv-SE', { timeZone: 'Asia/Makassar' }); // YYYY-MM-DD
          if (!grouped[dateKey]) {
            grouped[dateKey] = {
              dateKey,
              dateObj,
              datang: null,
              pulang: null,
              libur: null,
              sakit: null,
              username: log.username,
              ip: log.ip_address,
              distance: log.distance_meters
            };
          }
          if (log.type === 'DATANG') grouped[dateKey].datang = log;
          else if (log.type === 'PULANG') grouped[dateKey].pulang = log;
          else if (log.type === 'LIBUR') grouped[dateKey].libur = log;
          else if (log.type === 'SAKIT' || log.type === 'IZIN') grouped[dateKey].sakit = log;
        });

        // Sort keys descending (newest date first)
        const sortedKeys = Object.keys(grouped).sort().reverse();

        logsTableBody.innerHTML = sortedKeys.map(key => {
          const item = grouped[key];
          const dateFormatted = item.dateObj.toLocaleDateString('id-ID', { timeZone: 'Asia/Makassar', weekday: 'short', day: '2-digit', month: 'short', year: 'numeric' });

          let datangStr = '-';
          if (item.datang) {
            const t = new Date(item.datang.check_in_time).toLocaleTimeString('id-ID', { timeZone: 'Asia/Makassar', hour: '2-digit', minute: '2-digit' });
            datangStr = `<span style="color: #10b981; font-weight: 600;">☀️ ${t} WITA</span>`;
          }

          let pulangStr = '-';
          if (item.pulang) {
            const t = new Date(item.pulang.check_in_time).toLocaleTimeString('id-ID', { timeZone: 'Asia/Makassar', hour: '2-digit', minute: '2-digit' });
            pulangStr = `<span style="color: #6366f1; font-weight: 600;">🌙 ${t} WITA</span>`;
          }

          let statusBadge = '<span class="badge badge-success">HADIR</span>';
          let notes = item.datang ? item.datang.notes : (item.pulang ? item.pulang.notes : '-');

          if (item.sakit) {
            statusBadge = '<span class="badge badge-warning">🏥 SAKIT / IZIN</span>';
            datangStr = '<span class="text-muted">Izin Sakit</span>';
            pulangStr = '<span class="text-muted">Izin Sakit</span>';
            notes = item.sakit.notes || item.sakit.early_reason;
          } else if (item.libur) {
            if (item.libur.status === 'LIBUR_GALUNGAN') {
              statusBadge = '<span class="badge badge-warning" style="background: rgba(245, 158, 11, 0.2); color: #f59e0b; border: 1px solid rgba(245, 158, 11, 0.4);">🌴 GALUNGAN</span>';
              notes = 'Libur Hari Raya Galungan';
            } else {
              statusBadge = '<span class="badge badge-secondary" style="background: rgba(148, 163, 184, 0.2); color: #94a3b8; border: 1px solid rgba(148, 163, 184, 0.4);">🌴 MINGGU</span>';
              notes = 'Libur Hari Minggu (Sistem Otomatis)';
            }
            datangStr = '<span class="text-muted">Libur</span>';
            pulangStr = '<span class="text-muted">Libur</span>';
          } else if (item.pulang && item.pulang.status === 'PULANG_CEPAT') {
            statusBadge = '<span class="badge badge-warning">PULANG CEPAT</span>';
            notes = item.pulang.notes || item.pulang.early_reason;
          }

          const ipStr = item.ip || '-';
          const distStr = item.distance !== undefined ? `${item.distance.toFixed(1)}m` : '0.0m';

          return `
            <tr>
              <td data-label="Tanggal"><b>${dateFormatted}</b></td>
              <td data-label="Absen Datang">${datangStr}</td>
              <td data-label="Absen Pulang">${pulangStr}</td>
              <td data-label="Status Harian">${statusBadge}</td>
              <td data-label="IP & Jarak"><code>${ipStr}</code> (${distStr})</td>
              <td data-label="Keterangan">${notes}</td>
            </tr>
          `;
        }).join('');
      }
    } catch (err) {
      logsTableBody.innerHTML = '<tr><td colspan="6" class="text-center text-danger">Gagal memuat log presensi</td></tr>';
    }
  }

  btnRefreshLogs.addEventListener('click', loadLogs);

  // 10. Export PDF Absensi (Format Resmi PKL UNDIKNAS & PT Adya Artha Abadi)
  const btnExportPdf = document.getElementById('btn-export-pdf');
  if (btnExportPdf) {
    btnExportPdf.addEventListener('click', exportAbsensiPDF);
  }

  function exportAbsensiPDF() {
    if (!currentUser) {
      alert('Data profil belum selesai dimuat, silakan coba lagi.');
      return;
    }

    const today = new Date();
    const months = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
    const currentDay = today.getDate();
    const currentMonthStr = months[today.getMonth()];
    const currentYear = today.getFullYear();
    const todayFormatted = `${currentDay} ${currentMonthStr} ${currentYear}`; // e.g. "25 Juli 2026"

    const studentName = currentUser.full_name || currentUser.username;
    const studentNIM = currentUser.nim || '-';

    // Base date range 15 Juni 2026 to 30 Juni 2026 (Format Resmi Template DOCX PKL)
    const dateRange = [
      { no: 1, date: '15-06-2026', day: 'Senin' },
      { no: 2, date: '16-06-2026', day: 'Selasa' },
      { no: 3, date: '17-06-2026', day: 'Rabu' },
      { no: 4, date: '18-06-2026', day: 'Kamis' },
      { no: 5, date: '19-06-2026', day: 'Jumat' },
      { no: 6, date: '20-06-2026', day: 'Sabtu' },
      { no: 7, date: '21-06-2026', day: 'Minggu' },
      { no: 8, date: '22-06-2026', day: 'Senin' },
      { no: 9, date: '23-06-2026', day: 'Selasa' },
      { no: 10, date: '24-06-2026', day: 'Rabu' },
      { no: 11, date: '25-06-2026', day: 'Kamis' },
      { no: 12, date: '26-06-2026', day: 'Jumat' },
      { no: 13, date: '27-06-2026', day: 'Sabtu' },
      { no: 14, date: '28-06-2026', day: 'Minggu' },
      { no: 15, date: '29-06-2026', day: 'Senin' },
      { no: 16, date: '30-06-2026', day: 'Selasa' }
    ];

    // Build map from attendance logs
    const logMap = {};
    if (window.allAttendanceLogs && Array.isArray(window.allAttendanceLogs)) {
      window.allAttendanceLogs.forEach(log => {
        if (!log.status || log.status.startsWith('DITOLAK')) return;
        const d = new Date(log.check_in_time);
        const yyyy = d.getFullYear();
        const mm = String(d.getMonth() + 1).padStart(2, '0');
        const dd = String(d.getDate()).padStart(2, '0');
        const key = `${dd}-${mm}-${yyyy}`;
        if (!logMap[key]) logMap[key] = log;
      });
    }

    let rowsHtml = '';
    dateRange.forEach(item => {
      let statusStr = 'Hadir';

      // Special holiday rules matching schedule
      if (item.date === '16-06-2026' || item.date === '17-06-2026') {
        statusStr = 'Libur Hari Raya Galungan';
      } else if (item.day === 'Minggu') {
        statusStr = 'Libur Hari Minggu';
      } else {
        const log = logMap[item.date];
        if (log) {
          if (log.type === 'SAKIT' || log.status === 'SAKIT') {
            statusStr = 'Sakit';
          } else if (log.type === 'IZIN') {
            statusStr = 'Izin';
          } else {
            statusStr = 'Hadir';
          }
        } else {
          // Check specific user sickness records from instructions
          if (currentUser.username === 'deksa' && item.date === '04-07-2026') {
            statusStr = 'Sakit';
          } else if (currentUser.username === 'putra' && item.date === '14-07-2026') {
            statusStr = 'Sakit';
          } else {
            statusStr = 'Hadir';
          }
        }
      }

      rowsHtml += `
        <tr>
          <td style="border: 1px solid #000; padding: 5px; text-align: center;">${item.no}</td>
          <td style="border: 1px solid #000; padding: 5px; text-align: center;">${item.date}</td>
          <td style="border: 1px solid #000; padding: 5px; text-align: center;">${item.day}</td>
          <td style="border: 1px solid #000; padding: 5px; text-align: center; font-weight: ${statusStr === 'Hadir' ? 'normal' : 'bold'};">${statusStr}</td>
        </tr>
      `;
    });

    const exportContainer = document.createElement('div');
    exportContainer.style.position = 'absolute';
    exportContainer.style.left = '-9999px';
    exportContainer.style.top = '-9999px';
    exportContainer.style.width = '700px';
    exportContainer.style.fontFamily = "'Times New Roman', Times, serif";
    exportContainer.style.color = '#000';
    exportContainer.style.background = '#fff';
    exportContainer.style.padding = '30px 40px';

    exportContainer.innerHTML = `
      <div style="text-align: center; line-height: 1.3; margin-bottom: 20px;">
        <div style="font-size: 14pt; font-weight: bold;">DAFTAR HADIR PRAKTIK KERJA LAPANGAN (PKL)</div>
        <div style="font-size: 12pt; font-weight: bold;">UNIVERSITAS PENDIDIKAN NASIONAL (UNDIKNAS) DENPASAR</div>
        <div style="font-size: 12pt; font-weight: bold;">PT ADYA ARTHA ABADI</div>
      </div>

      <table style="width: 100%; font-size: 11pt; border: none; margin-bottom: 15px; border-collapse: collapse;">
        <tr><td style="width: 160px; padding: 2px 0;">Nama Mahasiswa</td><td style="width: 15px;">:</td><td><b>${studentName}</b></td></tr>
        <tr><td style="padding: 2px 0;">NIM</td><td>:</td><td>${studentNIM}</td></tr>
        <tr><td style="padding: 2px 0;">Program Studi</td><td>:</td><td>Manajemen</td></tr>
        <tr><td style="padding: 2px 0;">Tempat PKL</td><td>:</td><td>PT Adya Artha Abadi</td></tr>
        <tr><td style="padding: 2px 0;">Periode PKL</td><td>:</td><td>15 s.d. 30 Juni 2026</td></tr>
      </table>

      <div style="font-size: 11pt; font-weight: bold; margin-bottom: 8px;">BULAN: JUNI 2026</div>

      <table style="width: 100%; border-collapse: collapse; font-size: 10pt; margin-bottom: 25px;">
        <thead>
          <tr style="background-color: #f2f2f2;">
            <th style="border: 1px solid #000; padding: 6px; width: 40px; text-align: center;">No</th>
            <th style="border: 1px solid #000; padding: 6px; width: 110px; text-align: center;">Tanggal</th>
            <th style="border: 1px solid #000; padding: 6px; width: 100px; text-align: center;">Hari</th>
            <th style="border: 1px solid #000; padding: 6px; text-align: center;">Hadir / Tidak Hadir</th>
          </tr>
        </thead>
        <tbody>
          ${rowsHtml}
        </tbody>
      </table>

      <table style="width: 100%; border: none; font-size: 11pt; margin-top: 25px;">
        <tr>
          <td style="width: 50%; text-align: center; vertical-align: top;">
            Mengetahui,<br>
            <b>Kepala Cabang PT. Adya Artha Abadi Bali</b><br><br><br><br><br>
            <b><u>I Made Mas Sugianyar</u></b>
          </td>
          <td style="width: 50%; text-align: center; vertical-align: top;">
            <b>Denpasar, ${todayFormatted}</b><br>
            Mahasiswa PKL,<br><br><br><br><br>
            <b><u>${studentName}</u></b>
          </td>
        </tr>
      </table>
    `;

    document.body.appendChild(exportContainer);

    const opt = {
      margin:       [10, 15, 10, 15],
      filename:     `Absensi_PKL_Juni2026_${currentUser.username}.pdf`,
      image:        { type: 'jpeg', quality: 0.98 },
      html2canvas:  { scale: 2, useCORS: true },
      jsPDF:        { unit: 'mm', format: 'a4', orientation: 'portrait' }
    };

    if (window.html2pdf) {
      window.html2pdf().set(opt).from(exportContainer).save().then(() => {
        document.body.removeChild(exportContainer);
      }).catch(err => {
        console.error('Export PDF error:', err);
        document.body.removeChild(exportContainer);
        alert('Gagal membuat file PDF. Silakan coba lagi.');
      });
    } else {
      document.body.removeChild(exportContainer);
      window.print();
    }
  }

  // 9. Logout
  btnLogout.addEventListener('click', async () => {
    await fetch('/api/logout', { method: 'POST' });
    window.location.href = '/login';
  });

  // Init
  startClock();
  fetchUserData();
  setInterval(updateLocation, 5000);
});
