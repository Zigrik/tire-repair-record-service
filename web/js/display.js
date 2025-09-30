class QueueDisplay {
    constructor() {
        this.currentTimeElement = document.getElementById('currentTime');
        this.currentTickets = document.getElementById('currentTickets');
        this.queueDisplay = document.getElementById('queueDisplay');

        this.init();
    }

    init() {
        this.updateTime();
        setInterval(() => this.updateTime(), 1000);
        
        this.loadQueue();
        setInterval(() => this.loadQueue(), 3000); // Обновлять каждые 3 секунды

        console.log('QueueDisplay initialized');
    }

    updateTime() {
        const now = new Date();
        this.currentTimeElement.textContent = now.toLocaleString('ru-RU', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
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
        // Текущие в работе (максимум 3)
        const inWorkRecords = records.filter(record => 
            record.status === 'in work' || record.status === 'welcome'
        ).slice(0, 3);

        // Очередь ожидания
        const waitingRecords = records.filter(record => 
            record.status === 'wait'
        );

        // Отображаем текущие в работе
        if (inWorkRecords.length > 0) {
            this.currentTickets.innerHTML = inWorkRecords.map(record => `
                <div class="current-ticket">
                    <div class="ticket-number-large">${record.ticketNumber}</div>
                    <div class="car-number-large">${this.escapeHtml(record.title)}</div>
                    <div class="status-large">${this.getStatusText(record.status)}</div>
                </div>
            `).join('');
        } else {
            this.currentTickets.innerHTML = '<div class="no-tickets">НЕТ ЗАПИСЕЙ</div>';
        }

        // Отображаем очередь
        if (waitingRecords.length > 0) {
            this.queueDisplay.innerHTML = waitingRecords.map((record, index) => `
                <div class="queue-item-large">
                    <div class="queue-position">${index + 1}</div>
                    <div class="queue-ticket">${record.ticketNumber}</div>
                    <div class="queue-car">${this.escapeHtml(record.title)}</div>
                </div>
            `).join('');
        } else {
            this.queueDisplay.innerHTML = '<div class="no-queue">ОЧЕРЕДЬ ПУСТА</div>';
        }
    }

    getStatusText(status) {
        const statusMap = {
            'wait': 'ОЖИДАНИЕ',
            'welcome': 'ПРИНЯТ',
            'in work': 'В РАБОТЕ',
            'done': 'ЗАВЕРШЕН',
            'cancel': 'ОТМЕНЕН'
        };
        return statusMap[status] || status;
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new QueueDisplay();
});