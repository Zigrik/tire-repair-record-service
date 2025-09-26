// js/tire-service.js
class TireService {
    constructor() {
        this.currentQueue = [];
        this.inWork = [];
        this.updateInterval = null;
        this.autoRedirectTimer = null;
        
        this.initializeEventListeners();
        this.updateCurrentTime();
        this.loadQueue();
        this.startAutoUpdate();
    }

    initializeEventListeners() {
        // Переключение предварительной записи
        document.getElementById('preRecord').addEventListener('change', (e) => {
            document.getElementById('preRecordFields').style.display = e.target.checked ? 'block' : 'none';
            
            // Устанавливаем минимальную дату для записи (текущее время + 30 минут)
            if (e.target.checked) {
                const now = new Date();
                now.setMinutes(now.getMinutes() + 30);
                const minDate = now.toISOString().slice(0, 16);
                document.getElementById('recordDate').min = minDate;
            }
        });

        // Получение талона
        document.getElementById('getTicketBtn').addEventListener('click', () => {
            this.getTicket();
        });

        // Закрытие модального окна
        document.getElementById('closeModalBtn').addEventListener('click', () => {
            this.closeModal();
        });

        // Автоматическое закрытие модального окна по клику на фон
        document.getElementById('ticketModal').addEventListener('click', (e) => {
            if (e.target === document.getElementById('ticketModal')) {
                this.closeModal();
            }
        });

        // Enter для быстрого получения талона
        document.getElementById('carNumber').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.getTicket();
            }
        });
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
            // Загружаем все записи на сегодня
            const response = await axios.get('/api/GetTodayRecords');
            
            if (response.data.records) {
                this.processQueueData(response.data.records);
                this.renderQueue();
            }
        } catch (error) {
            console.error('Ошибка загрузки очереди:', error);
            this.showError('Ошибка загрузки очереди');
        }
    }

    processQueueData(records) {
        this.inWork = [];
        this.currentQueue = [];

        const now = new Date();
        
        // Фильтруем только активные записи (исключаем завершенные и отмененные)
        const activeRecords = records.filter(record => 
            record.status !== 'done' && record.status !== 'cancel'
        );

        // Разделяем записи по статусам
        activeRecords.forEach(record => {
            if (record.status === 'in work') {
                this.inWork.push(record);
            } else {
                this.currentQueue.push(record);
            }
        });

        // Сортируем очередь
        this.organizeQueue();
    }

    organizeQueue() {
        // Сначала обычная очередь (без предварительной записи)
        const regularQueue = this.currentQueue.filter(item => !item.record);
        
        // Затем предварительные записи, отсортированные по времени
        const appointments = this.currentQueue.filter(item => item.record)
            .sort((a, b) => new Date(a.record) - new Date(b.record));

        // Объединяем: сначала обычная очередь, потом предварительные записи
        this.currentQueue = [...regularQueue, ...appointments];
    }

    renderQueue() {
        this.renderInWork();
        this.renderQueueList();
    }

    renderInWork() {
        const container = document.getElementById('inWorkList');
        
        if (this.inWork.length === 0) {
            container.innerHTML = '<div class="empty-message">Нет машин в работе</div>';
            return;
        }

        container.innerHTML = this.inWork.map(record => `
            <div class="ticket-item current">
                <div class="ticket-info">
                    <div class="ticket-number">${this.formatTicketNumber(record)}</div>
                    <div class="ticket-car">${record.title}</div>
                    <div class="ticket-comment">${record.comment || ''}</div>
                </div>
                <div class="ticket-status">В работе</div>
            </div>
        `).join('');
    }

    renderQueueList() {
        const container = document.getElementById('queueList');
        
        if (this.currentQueue.length === 0) {
            container.innerHTML = '<div class="empty-message">Очередь пуста</div>';
            return;
        }

        container.innerHTML = this.currentQueue.map((record, index) => `
            <div class="ticket-item ${record.record ? 'record' : ''}">
                <div class="ticket-info">
                    <div class="ticket-number">${this.formatTicketNumber(record)}</div>
                    <div class="ticket-car">${record.title}</div>
                    <div class="ticket-comment">${record.comment || ''}</div>
                    ${record.record ? `<div class="ticket-time">Запись на: ${this.formatDateTime(record.record)}</div>` : ''}
                </div>
                <div class="ticket-position">${index + 1}</div>
            </div>
        `).join('');
    }

    formatTicketNumber(record) {
        const id = record.id ? record.id.toString().slice(-3).padStart(3, '0') : '000';
        return record.record ? `З${id}` : `О${id}`;
    }

    formatDateTime(dateTime) {
        const date = new Date(dateTime);
        return date.toLocaleString('ru-RU', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    async getTicket() {
        const carNumber = document.getElementById('carNumber').value.trim();
        const comment = document.getElementById('comment').value.trim();
        const isPreRecord = document.getElementById('preRecord').checked;
        const recordDate = document.getElementById('recordDate').value;

        if (!carNumber) {
            this.showError('Пожалуйста, введите номер автомобиля');
            return;
        }

        if (isPreRecord && !recordDate) {
            this.showError('Пожалуйста, выберите дату и время для предварительной записи');
            return;
        }

        // Проверка времени для предварительной записи
        if (isPreRecord) {
            const selectedTime = new Date(recordDate);
            const now = new Date();
            const minTime = new Date(now.getTime() + 30 * 60000); // +30 минут
            
            if (selectedTime < minTime) {
                this.showError('Время записи должно быть минимум на 30 минут позже текущего времени');
                return;
            }
        }

        try {
            const recordData = {
                title: carNumber,
                comment: comment,
                record: isPreRecord ? recordDate : null
            };

            console.log('Отправка данных:', recordData);
            
            const response = await axios.post('/api/AddRecord', recordData);
            
            console.log('Ответ сервера:', response.data);
            
            if (response.data.message === "Record added successfully") {
                // После успешного добавления загружаем обновленную очередь
                await this.loadQueue();
                
                // Показываем номер талона (используем последний ID из очереди)
                const lastRecord = this.currentQueue[this.currentQueue.length - 1];
                this.showTicket(lastRecord.id, carNumber, comment, isPreRecord);
                this.clearForm();
            } else {
                this.showError('Ошибка при получении талона');
            }
        } catch (error) {
            console.error('Ошибка получения талона:', error);
            if (error.response && error.response.data) {
                this.showError('Ошибка: ' + error.response.data.error);
            } else {
                this.showError('Ошибка при получении талона. Попробуйте еще раз.');
            }
        }
    }

    showTicket(recordId, carNumber, comment, isPreRecord) {
        const ticketNumber = this.formatTicketNumber({id: recordId, record: isPreRecord});
        
        document.getElementById('ticketNumber').textContent = ticketNumber;
        document.getElementById('ticketInfo').innerHTML = `
            <div><strong>Автомобиль:</strong> ${carNumber}</div>
            ${comment ? `<div><strong>Комментарий:</strong> ${comment}</div>` : ''}
            <div><strong>Тип:</strong> ${isPreRecord ? 'Предварительная запись' : 'Текущая очередь'}</div>
            <div class="redirect-timer">Автоматическое закрытие через: <span id="countdown">5</span> сек.</div>
        `;
        
        document.getElementById('ticketModal').style.display = 'flex';
        
        // Таймер обратного отсчета
        let countdown = 5;
        const countdownElement = document.getElementById('countdown');
        const countdownInterval = setInterval(() => {
            countdown--;
            countdownElement.textContent = countdown;
            
            if (countdown <= 0) {
                clearInterval(countdownInterval);
                this.closeModal();
            }
        }, 1000);
        
        // Сохраняем ID интервала для очистки
        this.autoRedirectTimer = countdownInterval;
    }

    closeModal() {
        document.getElementById('ticketModal').style.display = 'none';
        if (this.autoRedirectTimer) {
            clearInterval(this.autoRedirectTimer);
            this.autoRedirectTimer = null;
        }
    }

    clearForm() {
        document.getElementById('carNumber').value = '';
        document.getElementById('comment').value = '';
        document.getElementById('preRecord').checked = false;
        document.getElementById('preRecordFields').style.display = 'none';
        document.getElementById('recordDate').value = '';
        
        // Фокус на поле ввода номера автомобиля
        document.getElementById('carNumber').focus();
    }

    showError(message) {
        alert(message); // Можно заменить на красивый toast
    }

    startAutoUpdate() {
        this.updateInterval = setInterval(() => {
            this.loadQueue();
        }, 20000); // Обновление каждые 20 секунд
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    new TireService();
    
    // Фокус на поле ввода номера автомобиля
    document.getElementById('carNumber').focus();
});