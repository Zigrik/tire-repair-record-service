// js/display.js
class QueueDisplay {
    constructor() {
        this.queueData = [];
        this.updateInterval = null;
        
        this.initialize();
        this.loadQueue();
        this.startAutoUpdate();
    }

    initialize() {
        this.updateCurrentTime();
    }

    updateCurrentTime() {
        const updateTime = () => {
            const now = new Date();
            document.getElementById('currentTime').textContent = 
                now.toLocaleDateString('ru-RU') + ' ' + 
                now.toLocaleTimeString('ru-RU');
        };
        
        updateTime();
        setInterval(updateTime, 1000);
    }

    async loadQueue() {
        try {
            const response = await axios.get('/api/GetTodayRecords');
            this.queueData = response.data.records || [];
            this.renderQueue();
        } catch (error) {
            console.error('Ошибка загрузки очереди:', error);
        }
    }

    renderQueue() {
        this.renderCurrentTickets();
        this.renderQueueList();
    }

    renderCurrentTickets() {
        const container = document.getElementById('currentTickets');
        const inWork = this.queueData.filter(record => record.status === 'in work');
        
        if (inWork.length === 0) {
            container.innerHTML = `
                <div class="ticket-large">
                    <div class="ticket-number-large">---</div>
                    <div class="status">СВОБОДНО</div>
                </div>
            `;
            return;
        }

        container.innerHTML = inWork.map(record => `
            <div class="ticket-large current">
                <div class="ticket-number-large">${this.formatTicketNumber(record)}</div>
                <div class="car-number">${record.title}</div>
                <div class="status">В РАБОТЕ</div>
                <div class="start-time">С ${this.formatTime(record.date)}</div>
            </div>
        `).join('');
    }

    renderQueueList() {
        const container = document.getElementById('queueDisplay');
        const waiting = this.queueData.filter(record => 
            record.status === 'wait' || record.status === 'welcome'
        );

        // Сортируем очередь
        const sortedQueue = waiting.sort((a, b) => {
            if (a.record && b.record) return new Date(a.record) - new Date(b.record);
            if (a.record && !b.record) return -1;
            if (!a.record && b.record) return 1;
            return new Date(a.date) - new Date(b.date);
        });

        if (sortedQueue.length === 0) {
            container.innerHTML = '<div class="empty-queue">ОЧЕРЕДЬ ПУСТА</div>';
            return;
        }

        container.innerHTML = sortedQueue.map((record, index) => `
            <div class="ticket-item-large ${record.record ? 'appointment' : ''}">
                <div class="ticket-number-medium">${this.formatTicketNumber(record)}</div>
                <div class="position">${index + 1}</div>
                <div class="car-number">${record.title}</div>
                ${record.record ? `<div class="appointment-time">${this.formatTime(record.record)}</div>` : ''}
            </div>
        `).join('');
    }

    formatTicketNumber(record) {
        const id = record.id.toString().slice(-3).padStart(3, '0');
        return record.record ? `З${id}` : `О${id}`;
    }

    formatTime(dateTime) {
        const date = new Date(dateTime);
        return date.toLocaleTimeString('ru-RU', {
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    startAutoUpdate() {
        this.updateInterval = setInterval(() => {
            this.loadQueue();
        }, 10000); // Обновление каждые 10 секунд
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    new QueueDisplay();
});