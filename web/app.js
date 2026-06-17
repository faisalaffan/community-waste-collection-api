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
        { label: 'Households', value: 0, icon: 'ti ti-home', bg: 'bg-emerald-50', iconColor: 'text-emerald-600' },
        { label: 'Pickups', value: 0, icon: 'ti ti-truck', bg: 'bg-blue-50', iconColor: 'text-blue-600' },
        { label: 'Payments', value: 0, icon: 'ti ti-credit-card', bg: 'bg-amber-50', iconColor: 'text-amber-600' },
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
      editingPickup: null,
      pickupForm: { household_id: '', type: '', safety_check: false },
      pickupError: '',
      activeDropdownId: null,

      // Payments
      payments: [],

      // Proof modal
      showProofModal: false,
      proofUrl: '',

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
      const map = {
        pending: 'bg-amber-50 text-amber-700 border border-amber-100/70',
        scheduled: 'bg-blue-50 text-blue-700 border border-blue-100/70',
        completed: 'bg-emerald-50 text-emerald-700 border border-emerald-100/70',
        canceled: 'bg-rose-50 text-rose-700 border border-rose-100/70',
        paid: 'bg-emerald-50 text-emerald-700 border border-emerald-100/70'
      };
      return (map[s] || 'bg-slate-50 text-slate-600 border border-slate-100/70') + ' px-2.5 py-1 rounded-lg text-xs font-semibold inline-flex items-center gap-1 capitalize';
    },

    // ── Dashboard ──
    async loadDashboard() {
      const [w, p] = await Promise.all([this.api('/reports/waste-summary'), this.api('/reports/payment-summary')]);
      if (w.status === 'success') {
        const raw = w.data || [];
        const groups = {};
        const types = ['organic', 'plastic', 'paper', 'electronic'];
        types.forEach(t => {
          groups[t] = { type: t, total_count: 0, pending: 0, completed: 0 };
        });
        raw.forEach(item => {
          const t = item.type;
          if (!groups[t]) {
            groups[t] = { type: t, total_count: 0, pending: 0, completed: 0 };
          }
          groups[t].total_count += item.count || 0;
          if (item.status === 'pending') {
            groups[t].pending += item.count || 0;
          } else if (item.status === 'completed') {
            groups[t].completed += item.count || 0;
          }
        });
        this.wasteSummary = Object.values(groups);
      }
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
      let r;
      if (this.editingHousehold) {
        r = await this.api('/households/' + this.editingHousehold.id, { method: 'PUT', body });
      } else {
        r = await this.api('/households', { method: 'POST', body });
      }
      if (r.status === 'success') {
        this.showHouseholdForm = false;
        this.editingHousehold = null;
        this.householdForm = { owner_name: '', address: '' };
        this.loadHouseholds();
        this.loadDashboard();
      }
      else this.householdError = r.error?.message || 'Error';
    },
    editHousehold(h) {
      this.editingHousehold = h;
      this.householdForm = { owner_name: h.owner_name, address: h.address };
      this.showHouseholdForm = true;
      this.householdError = '';
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
      const body = JSON.stringify({ type: this.pickupForm.type, safety_check: this.pickupForm.safety_check });
      let r;
      if (this.editingPickup) {
        r = await this.api('/pickups/' + this.editingPickup.id, { method: 'PUT', body });
      } else {
        r = await this.api('/pickups', { method: 'POST', body: JSON.stringify(this.pickupForm) });
      }
      if (r.status === 'success') {
        this.showPickupForm = false;
        this.editingPickup = null;
        this.pickupForm = { household_id: '', type: '', safety_check: false };
        this.loadPickups();
        this.loadDashboard();
      }
      else this.pickupError = r.error?.message || 'Error';
    },
    editPickup(p) {
      this.editingPickup = p;
      this.pickupForm = { household_id: p.household_id, type: p.type, safety_check: p.safety_check };
      this.showPickupForm = true;
      this.pickupError = '';
    },
    async deletePickup(id) {
      if (!confirm('Delete this pickup?')) return;
      await this.api('/pickups/' + id, { method: 'DELETE' });
      this.loadPickups();
      this.loadDashboard();
    },
    async schedulePickup(id) {
      const date = prompt('Pickup date (YYYY-MM-DD):', new Date().toISOString().slice(0, 10));
      if (!date) return;
      await this.api('/pickups/' + id + '/schedule', { method: 'PUT', body: JSON.stringify({ pickup_date: date + 'T00:00:00Z' }) });
      this.loadPickups();
    },
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
    viewProof(paymentId) {
      this.proofUrl = '/api/files/proof/' + paymentId;
      this.showProofModal = true;
    },

    async loadHistory() {
      if (!this.historyHouseholdId) { this.history = null; return; }
      const r = await this.api('/reports/households/' + this.historyHouseholdId + '/history');
      if (r.status === 'success') this.history = r.data;
    },
    toggleDropdown(id) {
      this.activeDropdownId = this.activeDropdownId === id ? null : id;
    },
  },

  watch: {
    page(p) {
      localStorage.setItem('page', p);
      if (p === 'dashboard') this.loadDashboard();
      else if (p === 'households') this.loadHouseholds();
      else if (p === 'pickups') { this.loadPickups(); this.loadHouseholds(); }
      else if (p === 'payments') this.loadPayments();
      else if (p === 'reports') { this.loadDashboard(); this.loadHouseholds(); }
    },
    householdPage() { this.loadHouseholds(); },
  },

  mounted() {
    const saved = localStorage.getItem('page');
    if (saved && this.nav.find(n => n.id === saved)) this.page = saved;
    this.loadDashboard();

    // Tutup dropdown jika mengklik di luar area dropdown
    document.addEventListener('click', () => {
      this.activeDropdownId = null;
    });
  },
}).mount('#app');
