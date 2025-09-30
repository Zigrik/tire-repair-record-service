class AdminPanel {
    constructor() {
        this.currentTab = 'queue';
        this.currentMonth = new Date();
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadQueue();
        this.loadCalendar();

        console.log('AdminPanel initialized');
    }

    bindEvents() {
        // Переключение вкладок
        document.querySelectorAll('.tab-link').forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.switchTab(e.target.dataset.tab);
            });
        });

        // Кнопки управления очередью
        document.getElementById('nextBtn').addEventListener('click', () => this.nextCustomer());
        document.getElementById('cancelBtn').addEventListener('click', () => this.cancelCurrent());
        document.getElementById('logoutBtn').addEventListener('click', () => this.logout());

        // Фильтры записей
        document.getElementById('applyFilters').addEventListener('click', () => this.loadRecords());

        // Управление календарем
        document.getElementById('prevMonth').addEventListener('click', () => this.changeMonth(-1));
        document.getElementById('nextMonth').addEventListener('click', () => this.changeMonth(1));

        // Модальное окно
        document.getElementById('cancelEdit').addEventListener('click', () => this.closeModal());
        document.getElementById('saveEdit').addEventListener('click', () => this.saveRecord());
        document.getElementById('deleteRecord').addEventListener('click', () => this.deleteRecord());
    }

    switchTab(tabName) {
        this.currentTab = tabName;
        
        // Обновляем активные вкладки
        document.querySelectorAll('.tab-link').forEach(tab => {
            tab.classList.toggle('active', tab.dataset.tab === tabName);
        });
        
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.toggle('active', content.id === tabName + 'Tab');
        });

        // Загружаем данные для вкладки
        switch(tabName) {
            case 'queue':
                this.loadQueue();
                break;
            case 'records':
                this.loadRecords();
                break;
            case 'schedule':
                this.loadCalendar();
                break;
        }
    }

    async loadQueue() {
        try {
            const response = await axios.get('/api/GetTodayRecords');
            const records = response.data.records || [];
            this.displayQueue(records);
        } catch (error) {
            console.error('Ошибка загрузки очереди:', error);
        }
    }

    displayQueue(records) {
        const inWorkRecords = records.filter(record => 
            record.status === 'in work' || record.status === 'welcome'
        );
        
        const waitingRecords = records.filter(record => 
            record.status === 'wait'
        );

        // Текущие в работе
        const currentWorkList = document.getElementById('currentWorkList');
        if (inWorkRecords.length > 0) {
            currentWorkList.innerHTML = inWorkRecords.map(record => `
                <div class="record-card" data-id="${record.id}">
                    <div class="record-header">
                        <span class="ticket-number">${record.ticketNumber}</span>
                        <span class="car-number">${this.escapeHtml(record.title)}</span>
                        <span class="status-badge ${record.status}">${this.getStatusText(record.status)}</span>
                    </div>
                    <div class="record-details">
                        <div>Запись: ${this.formatDateTime(record.record)}</div>
                        <div>Комментарий: ${this.escapeHtml(record.comment || 'нет')}</div>
                    </div>
                    <div class="record-actions">
                        <button class="btn small" onclick="admin.editRecord(${record.id})">Редактировать</button>
                    </div>
                </div>
            `).join('');
        } else {
            currentWorkList.innerHTML = '<div class="empty-state">Нет записей в работе</div>';
        }

        // Очередь ожидания
        const waitingQueueList = document.getElementById('waitingQueueList');
        if (waitingRecords.length > 0) {
            waitingQueueList.innerHTML = waitingRecords.map(record => `
                <div class="record-card" data-id="${record.id}">
                    <div class="record-header">
                        <span class="ticket-number">${record.ticketNumber}</span>
                        <span class="car-number">${this.escapeHtml(record.title)}</span>
                    </div>
                    <div class="record-details">
                        <div>Запись: ${this.formatDateTime(record.record)}</div>
                        <div>Комментарий: ${this.escapeHtml(record.comment || 'нет')}</div>
                        <div>Создана: ${this.formatDateTime(record.date)}</div>
                    </div>
                    <div class="record-actions">
                        <button class="btn small primary" onclick="admin.updateStatus(${record.id}, 'welcome')">Принять</button>
                        <button class="btn small" onclick="admin.editRecord(${record.id})">Редактировать</button>
                    </div>
                </div>
            `).join('');
        } else {
            waitingQueueList.innerHTML = '<div class="empty-state">Очередь пуста</div>';
        }
    }

    // ... остальные методы для админки

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    formatDateTime(dateTime) {
        if (!dateTime) return 'Текущая очередь';
        const date = new Date(dateTime);
        return date.toLocaleString('ru-RU');
    }

    getStatusText(status) {
        const statusMap = {
            'wait': 'Ожидание',
            'welcome': 'Принят',
            'in work': 'В работе',
            'done': 'Завершен',
            'cancel': 'Отменен'
        };
        return statusMap[status] || status;
    }
}

// Глобальный экземпляр для обработки событий
const admin = new AdminPanel();