const { createApp } = Vue;

createApp({
  data() {
    return {
      page: 'dashboard',
      mobileOpen: false,
      nav: [
        { id: 'dashboard', label: 'Dashboard', icon: 'ti ti-layout-dashboard' },
        { id: 'households', label: 'Households', icon: 'ti ti-home' },
        { id: 'pickups', label: 'Pickups', icon: 'ti ti-truck' },
        { id: 'payments', label: 'Payments', icon: 'ti ti-credit-card' },
        { id: 'reports', label: 'Reports', icon: 'ti ti-chart-bar' },
      ],

      // Dashboard
      cards: [
        { label: 'Households', value: 0, icon: 'ti ti-home', bg: 'bg-green-50', iconColor: 'text-green-600' },
        { label: 'Pickups', value: 0, icon: 'ti ti-truck', bg: 'bg-blue-50', iconColor: 'text-blue-600' },
        { label: 'Payments', value: 0, icon: 'ti ti-credit-card', bg: 'bg-yellow-50', iconColor: 'text-yellow-600' },
        { label: 'Revenue', value: 'Rp 0', icon: 'ti ti-cash', bg: 'bg-purple-50', iconColor: 'text-purple-600' },
      ],
      wasteSummary: null,
      paymentSummary: null,

      // Households
      households: [],
      householdPage: 1,
      showHouseholdForm: false,
      editingHousehold: null,
      householdForm: { owner_name: '', address: '' },
      householdError: '',

      // Pickups
      pickups: [],
      showPickupForm: false,
      pickupForm: { household_id: '', type: '', safety_check: false },
      pickupError: '',

      // Payments
      payments: [],

      // Reports
      historyHouseholdId: '',
      history: null,
    };
  },

  methods: {
    api(path, opts = {}) { return fetch('/api' + path, { headers: { 'Content-Type': 'application/json' }, ...opts }).then(r => r.json()); },
    fmt(n) { return n ? Number(n).toLocaleString('id-ID') : '0'; },
    d(s) { return s ? new Date(s).toLocaleDateString('id-ID') : ''; },
    statusBadge(s) {
      const map = { pending: 'bg-yellow-100 text-yellow-700', scheduled: 'bg-blue-100 text-blue-700', completed: 'bg-green-100 text-green-700', canceled: 'bg-red-100 text-red-700', paid: 'bg-green-100 text-green-700' };
      return map[s] || 'bg-gray-100 text-gray-600';
    },

    // ── Dashboard ──
    async loadDashboard() {
      const [w, p] = await Promise.all([this.api('/reports/waste-summary'), this.api('/reports/payment-summary')]);
      if (w.status === 'success') this.wasteSummary = w.data;
      if (p.status === 'success') {
        this.paymentSummary = p.data;
        this.cards[3].value = 'Rp ' + this.fmt(p.data.total_revenue);
      }
      const hh = await this.api('/households?page=1&per_page=1');
      if (hh.status === 'success') this.cards[0].value = hh.pagination?.total || 0;
      const pp = await this.api('/pickups?page=1&per_page=1');
      if (pp.status === 'success') this.cards[1].value = pp.pagination?.total || 0;
      const pm = await this.api('/payments?page=1&per_page=1');
      if (pm.status === 'success') this.cards[2].value = pm.pagination?.total || 0;
    },

    // ── Households ──
    async loadHouseholds() {
      const r = await this.api('/households?page=' + this.householdPage + '&per_page=10');
      if (r.status === 'success') this.households = r.data;
    },
    async saveHousehold() {
      this.householdError = '';
      const body = JSON.stringify(this.householdForm);
      const r = await this.api('/households', { method: 'POST', body });
      if (r.status === 'success') { this.showHouseholdForm = false; this.householdForm = { owner_name: '', address: '' }; this.loadHouseholds(); this.loadDashboard(); }
      else this.householdError = r.error?.message || 'Error';
    },
    async deleteHousehold(id) {
      if (!confirm('Delete this household?')) return;
      await this.api('/households/' + id, { method: 'DELETE' });
      this.loadHouseholds(); this.loadDashboard();
    },

    // ── Pickups ──
    async loadPickups() {
      const r = await this.api('/pickups?per_page=50');
      if (r.status === 'success') this.pickups = r.data;
    },
    async savePickup() {
      this.pickupError = '';
      const body = JSON.stringify(this.pickupForm);
      const r = await this.api('/pickups', { method: 'POST', body });
      if (r.status === 'success') { this.showPickupForm = false; this.pickupForm = { household_id: '', type: '', safety_check: false }; this.loadPickups(); this.loadDashboard(); }
      else this.pickupError = r.error?.message || 'Error';
    },
    async schedulePickup(id) { await this.api('/pickups/' + id + '/schedule', { method: 'PUT' }); this.loadPickups(); },
    async completePickup(id) { await this.api('/pickups/' + id + '/complete', { method: 'PUT' }); this.loadPickups(); this.loadDashboard(); },
    async cancelPickup(id) { await this.api('/pickups/' + id + '/cancel', { method: 'PUT' }); this.loadPickups(); },

    // ── Payments ──
    async loadPayments() {
      const r = await this.api('/payments?per_page=50');
      if (r.status === 'success') this.payments = r.data;
    },
    async confirmPayment(id, event) {
      const file = event.target.files[0];
      if (!file) return;
      const form = new FormData();
      form.append('proof_file', file);
      await fetch('/api/payments/' + id + '/confirm', { method: 'PUT', body: form });
      this.loadPayments(); this.loadDashboard();
    },

    // ── Reports ──
    async loadHistory() {
      if (!this.historyHouseholdId) { this.history = null; return; }
      const r = await this.api('/reports/households/' + this.historyHouseholdId + '/history');
      if (r.status === 'success') this.history = r.data;
    },
  },

  watch: {
    page(p) {
      if (p === 'dashboard') this.loadDashboard();
      else if (p === 'households') this.loadHouseholds();
      else if (p === 'pickups') { this.loadPickups(); this.loadHouseholds(); }
      else if (p === 'payments') this.loadPayments();
      else if (p === 'reports') { this.loadDashboard(); this.loadHouseholds(); }
    },
    householdPage() { this.loadHouseholds(); },
  },

  mounted() {
    this.loadDashboard();
  },
}).mount('#app');
