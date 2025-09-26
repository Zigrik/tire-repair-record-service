// js/admin.js
class AdminPanel {
    constructor() {
        this.currentTab = 'queue';
        this.records = [];
        this.currentDate = new Date();
        
        this.initializeEventListeners();
        this.loadQueueData();
        this.switchTab('queue');
    }

    initializeEventListeners() {
        // Переключение вкладок
        document.querySelectorAll('.tab-link').forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.switchTab(e.target.dataset.tab);
            });
        });

        // Управление очередью
        document.getElementById('nextBtn').addEventListener('click', () => this.nextCustomer());
        document.getElementById('cancelBtn').addEventListener('click', () => this.cancelCurrent());

        // Фильтры записей
        document.getElementById('applyFilters').addEventListener('click', () => this.loadRecords());

        // Навигация по календарю
        document.getElementById('prevMonth').addEventListener('click', () => this.prevMonth());
        document.getElementById('nextMonth').addEventListener('click', () => this.nextMonth());

        // Редактирование записей
        document.getElementById('saveEdit').addEventListener('click', () => this.saveRecord());
        document.getElementById('cancelEdit').addEventListener('click', () => this.closeEditModal());
        document.getElementById('deleteRecord').addEventListener('click', () => this.deleteRecord());

        // Выход
        document.getElementById('logoutBtn').addEventListener('click', () => this.logout());
    }

    switchTab(tabName) {
        // Скрыть все вкладки
        document.querySelectorAll('.tab-content').forEach(tab => {
            tab.classList.remove('active');
        });
        document.querySelectorAll('.tab-link').forEach(tab => {
            tab.classList.remove('active');
        });

        // Показать выбранную вкладку
        document.getElementById(tabName + 'Tab').classList.add('active');
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        this.currentTab = tabName;

        // Загрузить данные для вкладки
        switch(tabName) {
            case 'queue':
                this.loadQueueData();
                break;
            case 'records':
                this.loadRecords();
                break;
            case 'schedule':
                this.renderCalendar();
                break;
        }
    }

    async loadQueueData() {
        try {
            const response = await axios.get('/api/GetTodayRecords');
            this.records = response.data.records || [];
            this.renderQueueManagement();
        } catch (error) {
            console.error('Ошибка загрузки очереди:', error);
            this.showError('Ошибка загрузки данных');
        }
    }

    renderQueueManagement() {
        const inWork = this.records.filter(r => r.status === 'in work');
        const waiting = this.records.filter(r => r.status === 'wait' || r.status === 'welcome');

        this.renderCurrentWork(inWork);
        this.renderWaitingQueue(waiting);
    }

    renderCurrentWork(records) {
        const container = document.getElementById('currentWorkList');
        
        if (records.length === 0) {
            container.innerHTML = '<div class="empty-message">Нет машин в работе</div>';
            return;
        }

        container.innerHTML = records.map(record => `
            <div class="work-item">
                <div class="work-info">
                    <div class="ticket-number">${this.formatTicketNumber(record)}</div>
                    <div class="car-number">${record.title}</div>
                    <div class="record-time">Начало: ${this.formatDateTime(record.date)}</div>
                </div>
                <div class="work-actions">
                    <button class="btn small success" onclick="admin.completeWork(${record.id})">Завершить</button>
                    <button class="btn small secondary" onclick="admin.editRecord(${record.id})">Редактировать</button>
                </div>
            </div>
        `).join('');
    }

    renderWaitingQueue(records) {
        const container = document.getElementById('waitingQueueList');
        
        if (records.length === 0) {
            container.innerHTML = '<div class="empty-message">Очередь пуста</div>';
            return;
        }

        // Сортируем: сначала по записи, потом по времени создания
        const sortedRecords = records.sort((a, b) => {
            if (a.record && b.record) return new Date(a.record) - new Date(b.record);
            if (a.record && !b.record) return -1;
            if (!a.record && b.record) return 1;
            return new Date(a.date) - new Date(b.date);
        });

        container.innerHTML = sortedRecords.map((record, index) => `
            <div class="queue-item ${record.record ? 'appointment' : ''}">
                <div class="queue-position">${index + 1}</div>
                <div class="queue-info">
                    <div class="ticket-number">${this.formatTicketNumber(record)}</div>
                    <div class="car-number">${record.title}</div>
                    ${record.record ? `<div class="appointment-time">Запись: ${this.formatDateTime(record.record)}</div>` : ''}
                    <div class="comment">${record.comment || ''}</div>
                </div>
                <div class="queue-actions">
                    <button class="btn small primary" onclick="admin.startWork(${record.id})">Принять</button>
                    <button class="btn small secondary" onclick="admin.editRecord(${record.id})">Редактировать</button>
                    <button class="btn small danger" onclick="admin.cancelRecord(${record.id})">Отменить</button>
                </div>
            </div>
        `).join('');
    }

    async startWork(recordId) {
        try {
            await axios.put('/api/UpdateRecordStatus', {
                id: recordId,
                status: 'in work'
            });
            this.loadQueueData();
            this.showSuccess('Машина принята в работу');
        } catch (error) {
            console.error('Ошибка:', error);
            this.showError('Ошибка при принятии машины');
        }
    }

    async completeWork(recordId) {
        try {
            await axios.put('/api/UpdateRecordStatus', {
                id: recordId,
                status: 'done'
            });
            this.loadQueueData();
            this.showSuccess('Работа завершена');
        } catch (error) {
            console.error('Ошибка:', error);
            this.showError('Ошибка при завершении работы');
        }
    }

    async cancelRecord(recordId) {
        if (confirm('Вы уверены, что хотите отменить эту запись?')) {
            try {
                await axios.put('/api/UpdateRecordStatus', {
                    id: recordId,
                    status: 'cancel'
                });
                this.loadQueueData();
                this.showSuccess('Запись отменена');
            } catch (error) {
                console.error('Ошибка:', error);
                this.showError('Ошибка при отмене записи');
            }
        }
    }

    async nextCustomer() {
        // Находим следующую запись в очереди
        const waiting = this.records.filter(r => r.status === 'wait' || r.status === 'welcome')
                                  .sort((a, b) => new Date(a.date) - new Date(b.date));
        
        if (waiting.length === 0) {
            this.showInfo('Нет машин в очереди');
            return;
        }

        const nextRecord = waiting[0];
        await this.startWork(nextRecord.id);
    }

    async cancelCurrent() {
        const inWork = this.records.filter(r => r.status === 'in work');
        
        if (inWork.length === 0) {
            this.showInfo('Нет машин в работе');
            return;
        }

        if (confirm('Отменить текущую работу?')) {
            for (const record of inWork) {
                await this.cancelRecord(record.id);
            }
        }
    }

    // Редактирование записей
    async editRecord(recordId) {
        try {
            const response = await axios.get(`/api/GetRecordByID?id=${recordId}`);
            const record = response.data.record;
            
            document.getElementById('editId').value = record.id;
            document.getElementById('editCarNumber').value = record.title;
            document.getElementById('editComment').value = record.comment || '';
            document.getElementById('editStatus').value = record.status;
            
            if (record.record) {
                const recordDate = new Date(record.record);
                document.getElementById('editRecordDate').value = recordDate.toISOString().slice(0, 16);
            } else {
                document.getElementById('editRecordDate').value = '';
            }
            
            document.getElementById('editModal').style.display = 'flex';
        } catch (error) {
            console.error('Ошибка:', error);
            this.showError('Ошибка загрузки данных записи');
        }
    }

    async saveRecord() {
        const recordData = {
            id: parseInt(document.getElementById('editId').value),
            title: document.getElementById('editCarNumber').value.trim(),
            comment: document.getElementById('editComment').value.trim(),
            status: document.getElementById('editStatus').value,
            record: document.getElementById('editRecordDate').value || null
        };

        if (!recordData.title) {
            this.showError('Введите номер автомобиля');
            return;
        }

        try {
            await axios.put('/api/UpdateRecord', recordData);
            this.closeEditModal();
            this.loadQueueData();
            this.showSuccess('Запись обновлена');
        } catch (error) {
            console.error('Ошибка:', error);
            this.showError('Ошибка при обновлении записи');
        }
    }

    async deleteRecord() {
        const recordId = document.getElementById('editId').value;
        
        if (confirm('Вы уверены, что хотите удалить эту запись?')) {
            try {
                await axios.delete(`/api/DeleteRecord?id=${recordId}`);
                this.closeEditModal();
                this.loadQueueData();
                this.showSuccess('Запись удалена');
            } catch (error) {
                console.error('Ошибка:', error);
                this.showError('Ошибка при удалении записи');
            }
        }
    }

    closeEditModal() {
        document.getElementById('editModal').style.display = 'none';
    }

    // Вспомогательные методы
    formatTicketNumber(record) {
        const id = record.id.toString().slice(-3).padStart(3, '0');
        return record.record ? `З${id}` : `О${id}`;
    }

    formatDateTime(dateTime) {
        const date = new Date(dateTime);
        return date.toLocaleString('ru-RU');
    }

    showError(message) {
        alert('Ошибка: ' + message);
    }

    showSuccess(message) {
        alert('Успех: ' + message);
    }

    showInfo(message) {
        alert('Информация: ' + message);
    }

    logout() {
        document.cookie = "token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
        window.location.href = '/login.html';
    }
}

// Глобальный экземпляр для обработки событий
const admin = new AdminPanel();