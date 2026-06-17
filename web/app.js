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
      householdPagination: null,

      // Pickups
      pickups: [],
      pickupPage: 1,
      pickupTotal: 0,
      showPickupForm: false,
      editingPickup: null,
      pickupForm: { household_id: '', type: '', safety_check: false },
      pickupError: '',
      activeDropdownId: null,
      pickupSearchQuery: '',
      showPickupAutocomplete: false,
      showScheduleModal: false,
      schedulePickupId: null,
      scheduleDate: '',
      calendarYear: new Date().getFullYear(),
      calendarMonth: new Date().getMonth(),

      // Payments
      payments: [],
      paymentPage: 1,
      paymentTotal: 0,

      // Proof modal
      showProofModal: false,
      proofUrl: '',

      // Reports
      historyHouseholdId: 'all',
      history: null,
      historySearchQuery: 'Semua Warga',
      showHistoryAutocomplete: false,
      allHouseholds: [],
    };
  },

  computed: {
    filteredHistoryHouseholds() {
      const q = (this.historySearchQuery || '').toLowerCase().trim();
      if (!q || q === 'semua warga') return this.allHouseholds;
      return this.allHouseholds.filter(h => h.owner_name.toLowerCase().includes(q));
    },
    filteredPickupHouseholds() {
      const q = (this.pickupSearchQuery || '').toLowerCase().trim();
      if (!q) return this.allHouseholds;
      return this.allHouseholds.filter(h => h.owner_name.toLowerCase().includes(q));
    },
    calendarDays() {
      if (this.calendarYear === undefined || this.calendarMonth === undefined) {
        return [];
      }
      const year = this.calendarYear;
      const month = this.calendarMonth;

      const firstDay = new Date(year, month, 1);
      let startDayOfWeek = firstDay.getDay();
      // Sun = 0, Mon = 1, Tue = 2, Wed = 3, Thu = 4, Fri = 5, Sat = 6
      // We want Monday (1) to be index 0, Tuesday (2) to be 1, ..., Sunday (0) to be 6.
      startDayOfWeek = startDayOfWeek === 0 ? 6 : startDayOfWeek - 1;

      const totalDays = new Date(year, month + 1, 0).getDate();
      const prevMonthTotalDays = new Date(year, month, 0).getDate();

      const days = [];

      // Prev month filler days
      for (let i = startDayOfWeek - 1; i >= 0; i--) {
        const d = prevMonthTotalDays - i;
        const prevMonth = month === 0 ? 11 : month - 1;
        const prevYear = month === 0 ? year - 1 : year;
        days.push({
          day: d,
          month: prevMonth,
          year: prevYear,
          isCurrentMonth: false,
          dateStr: `${prevYear}-${String(prevMonth + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
        });
      }

      // Current month days
      for (let d = 1; d <= totalDays; d++) {
        days.push({
          day: d,
          month: month,
          year: year,
          isCurrentMonth: true,
          dateStr: `${year}-${String(month + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
        });
      }

      // Next month filler days
      const remaining = 42 - days.length;
      for (let d = 1; d <= remaining; d++) {
        const nextMonth = month === 11 ? 0 : month + 1;
        const nextYear = month === 11 ? year + 1 : year;
        days.push({
          day: d,
          month: nextMonth,
          year: nextYear,
          isCurrentMonth: false,
          dateStr: `${nextYear}-${String(nextMonth + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
        });
      }

      return days;
    },
    calendarMonthName() {
      const names = [
        'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
        'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
      ];
      return names[this.calendarMonth] + ' ' + this.calendarYear;
    },
    todayDateStr() {
      const today = new Date();
      return today.getFullYear() + '-' + String(today.getMonth() + 1).padStart(2, '0') + '-' + String(today.getDate()).padStart(2, '0');
    }
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
        const order = ['organic', 'plastic', 'paper', 'electronic'];
        this.wasteSummary = order.map(t => {
          const found = raw.find(item => item.type === t);
          return found ? {
            type: t,
            total_count: found.total_count || 0,
            pending: found.pending || 0,
            completed: found.completed || 0,
            canceled: found.canceled || 0
          } : { type: t, total_count: 0, pending: 0, completed: 0, canceled: 0 };
        });
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
      if (r.status === 'success') {
        this.households = r.data;
        this.householdPagination = r.pagination;
      }
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
      const r = await this.api('/pickups?page=' + this.pickupPage + '&per_page=10');
      if (r.status === 'success') { this.pickups = r.data; this.pickupTotal = r.pagination?.total || 0; }
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
    openPickupForm() {
      this.pickupForm = { household_id: '', type: '', safety_check: false };
      this.pickupSearchQuery = '';
      this.showPickupAutocomplete = false;
      this.editingPickup = null;
      this.pickupError = '';
      this.showPickupForm = true;
    },
    selectPickupHousehold(id, name) {
      this.pickupForm.household_id = id;
      this.pickupSearchQuery = name;
      this.showPickupAutocomplete = false;
    },
    clearPickupSearch() {
      this.pickupForm.household_id = '';
      this.pickupSearchQuery = '';
      this.showPickupAutocomplete = false;
    },
    editPickup(p) {
      this.editingPickup = p;
      this.pickupForm = { household_id: p.household_id, type: p.type, safety_check: p.safety_check };
      const hh = this.allHouseholds.find(h => h.id === p.household_id);
      this.pickupSearchQuery = hh ? hh.owner_name : '';
      this.showPickupForm = true;
      this.pickupError = '';
    },
    async deletePickup(id) {
      if (!confirm('Delete this pickup?')) return;
      await this.api('/pickups/' + id, { method: 'DELETE' });
      this.loadPickups();
      this.loadDashboard();
    },
    openScheduleModal(id) {
      this.schedulePickupId = id;
      const today = new Date();
      this.scheduleDate = today.getFullYear() + '-' + String(today.getMonth() + 1).padStart(2, '0') + '-' + String(today.getDate()).padStart(2, '0');
      this.calendarYear = today.getFullYear();
      this.calendarMonth = today.getMonth();
      this.showScheduleModal = true;
    },
    prevMonth() {
      if (this.calendarMonth === 0) {
        this.calendarMonth = 11;
        this.calendarYear--;
      } else {
        this.calendarMonth--;
      }
    },
    nextMonth() {
      if (this.calendarMonth === 11) {
        this.calendarMonth = 0;
        this.calendarYear++;
      } else {
        this.calendarMonth++;
      }
    },
    selectCalendarDate(dayObj) {
      this.scheduleDate = dayObj.dateStr;
      if (dayObj.month !== this.calendarMonth) {
        this.calendarMonth = dayObj.month;
        this.calendarYear = dayObj.year;
      }
    },
    async saveSchedule() {
      if (!this.scheduleDate) return;
      await this.api('/pickups/' + this.schedulePickupId + '/schedule', {
        method: 'PUT',
        body: JSON.stringify({ pickup_date: this.scheduleDate + 'T00:00:00Z' })
      });
      this.showScheduleModal = false;
      this.schedulePickupId = null;
      this.loadPickups();
      this.loadDashboard();
    },
    async completePickup(id) { await this.api('/pickups/' + id + '/complete', { method: 'PUT' }); this.loadPickups(); this.loadDashboard(); },
    async cancelPickup(id) { await this.api('/pickups/' + id + '/cancel', { method: 'PUT' }); this.loadPickups(); },

    // ── Payments ──
    async loadPayments() {
      const r = await this.api('/payments?page=' + this.paymentPage + '&per_page=10');
      if (r.status === 'success') { this.payments = r.data; this.paymentTotal = r.pagination?.total || 0; }
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
    async loadAllHouseholds() {
      const r = await this.api('/households?per_page=1000');
      if (r.status === 'success') this.allHouseholds = r.data || [];
    },
    selectHistoryHousehold(id, name) {
      this.historyHouseholdId = id;
      this.historySearchQuery = name;
      this.showHistoryAutocomplete = false;
      this.loadHistory();
    },
    clearHistorySearch() {
      this.historyHouseholdId = 'all';
      this.historySearchQuery = 'Semua Warga';
      this.showHistoryAutocomplete = false;
      this.loadHistory();
    },
    toggleDropdown(id) {
      this.activeDropdownId = this.activeDropdownId === id ? null : id;
    },
    getInitials(name) {
      if (!name) return 'RT';
      const parts = name.trim().split(/\s+/);
      if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
      return (parts[0][0] + (parts[1] ? parts[1][0] : '')).toUpperCase();
    },
  },

  watch: {
    page(p) {
      localStorage.setItem('page', p);
      if (p === 'dashboard') this.loadDashboard();
      else if (p === 'households') this.loadHouseholds();
      else if (p === 'pickups') { this.loadPickups(); this.loadAllHouseholds(); }
      else if (p === 'payments') this.loadPayments();
      else if (p === 'reports') { this.loadDashboard(); this.loadAllHouseholds(); this.loadHistory(); }
    },
    householdPage() { this.loadHouseholds(); },
    pickupPage() { this.loadPickups(); },
    paymentPage() { this.loadPayments(); },
  },

  mounted() {
    const saved = localStorage.getItem('page');
    if (saved && this.nav.find(n => n.id === saved)) this.page = saved;
    this.loadDashboard();
    this.loadAllHouseholds();
    if (this.page === 'reports') this.loadHistory();

    // Tutup dropdown jika mengklik di luar area dropdown
    document.addEventListener('click', (e) => {
      this.activeDropdownId = null;
      if (!e.target.closest('.autocomplete-container')) {
        this.showHistoryAutocomplete = false;
      }
      if (!e.target.closest('.pickup-autocomplete-container')) {
        this.showPickupAutocomplete = false;
      }
    });
  },
}).mount('#app');
