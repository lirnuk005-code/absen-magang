document.addEventListener('DOMContentLoaded', () => {
  const loginForm = document.getElementById('login-form');

  // Check if user is already logged in
  fetch('/api/me')
    .then(res => res.json())
    .then(data => {
      if (data.success) {
        window.location.href = '/dashboard';
      }
    })
    .catch(() => {});

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
        window.location.href = '/dashboard';
      } else {
        alert(json.message || 'Login gagal!');
      }
    } catch (err) {
      alert('Gagal terhubung ke server!');
    }
  });
});
