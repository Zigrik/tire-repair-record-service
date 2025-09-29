class TireService {
    constructor() {
        this.currentTimeElement = document.getElementById('currentTime');
        this.inWorkList = document.getElementById('inWorkList');
        this.queueList = document.getElementById('queueList');
        this.carNumberInput = document.getElementById('carNumber');
        this.commentInput = document.getElementById('comment');
        this.preRecordCheckbox = document.getElementById('preRecord');
        this.preRecordFields = document.getElementById('preRecordFields');
        this.recordDateInput = document.getElementById('recordDate');
        this.getTicketBtn = document.getElementById('getTicketBtn');
        this.ticketModal = document.getElementById('ticketModal');
        this.ticketNumber = document.getElementById('ticketNumber');
        this.ticketInfo = document.getElementById('ticketInfo');
        this.closeModalBtn = document.getElementById('closeModalBtn');

        this.init();
    }

    init() {
        this.updateTime();
        setInterval(() => this.updateTime(), 1000);
        
        this.loadQueue();
        setInterval(() => this.loadQueue(), 5000);

        this.preRecordCheckbox.addEventListener('change', () => this.togglePreRecord());
        this.getTicketBtn.addEventListener('click', () => this.getTicket());
        this.closeModalBtn.addEventListener('click', () => this.closeModal());

        // Устанавливаем минимальную дату для записи
        const minDate = new Date();
        minDate.setMinutes(minDate.getMinutes() + 30);
        this.recordDateInput.min = minDate.toISOString().slice(0, 16);
        
        console.log('TireService initialized');
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

    togglePreRecord() {
        if (this.preRecordCheckbox.checked) {
            this.preRecordFields.style.display = 'block';
        } else {
            this.preRecordFields.style.display = 'none';
            this.recordDateInput.value = '';
        }
    }

    async loadQueue() {
        try {
            console.log('Loading queue...');
            const response = await axios.get('/api/GetTodayRecords');
            console.log('Queue response:', response.data);
            
            // Нормализуем поля - преобразуем заглавные в строчные
            const records = this.normalizeRecords(response.data.records || []);
            
            console.log('Normalized records:', records);
            this.displayQueue(records);
        } catch (error) {
            console.error('Ошибка загрузки очереди:', error);
            this.showError('Ошибка загрузки данных');
        }
    }

    // Преобразуем поля из заглавных в строчные
    normalizeRecords(records) {
        return records.map(record => ({
            id: record.ID || record.id,
            date: record.Date || record.date,
            title: record.Title || record.title,
            record: record.Record || record.record,
            comment: record.Comment || record.comment,
            status: record.Status || record.status
        }));
    }

    displayQueue(records) {
        console.log('Displaying records:', records);
        
        // Фильтруем записи по статусам
        const inWorkRecords = records.filter(record => 
            record.status === 'in work' || record.status === 'welcome'
        );
        
        const waitingRecords = records.filter(record => 
            record.status === 'wait'
        );

        console.log('In work records:', inWorkRecords);
        console.log('Waiting records:', waitingRecords);

        // Отображаем "В работе"
        if (inWorkRecords.length > 0) {
            this.inWorkList.innerHTML = inWorkRecords.map(record => `
                <div class="queue-item in-work">
                    <div class="car-number">${this.escapeHtml(record.title)}</div>
                    <div class="record-info">
                        <div class="record-time">${this.formatRecordTime(record.record)}</div>
                        <div class="status">${this.getStatusText(record.status)}</div>
                    </div>
                    ${record.comment ? `<div class="comment">Комментарий: ${this.escapeHtml(record.comment)}</div>` : ''}
                </div>
            `).join('');
        } else {
            this.inWorkList.innerHTML = '<div class="empty-message">Нет машин в работе</div>';
        }

        // Отображаем "Очередь"
        if (waitingRecords.length > 0) {
            this.queueList.innerHTML = waitingRecords.map((record, index) => `
                <div class="queue-item waiting">
                    <div class="car-number">${this.escapeHtml(record.title)}</div>
                    <div class="record-info">
                        <div class="record-time">${this.formatRecordTime(record.record)}</div>
                        <div class="position">Позиция: #${index + 1}</div>
                    </div>
                    ${record.comment ? `<div class="comment">Комментарий: ${this.escapeHtml(record.comment)}</div>` : ''}
                </div>
            `).join('');
        } else {
            this.queueList.innerHTML = '<div class="empty-message">Очередь пуста</div>';
        }
    }

    formatRecordTime(recordTime) {
        if (!recordTime) return 'Текущая очередь';
        
        try {
            const date = new Date(recordTime);
            return `Запись на: ${date.toLocaleString('ru-RU', {
                day: '2-digit',
                month: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            })}`;
        } catch (e) {
            return 'Текущая очередь';
        }
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

    async getTicket() {
        const carNumber = this.carNumberInput.value.trim();
        const comment = this.commentInput.value.trim();
        const isPreRecord = this.preRecordCheckbox.checked;
        const recordDate = isPreRecord ? this.recordDateInput.value : null;

        if (!carNumber) {
            alert('Пожалуйста, введите номер автомобиля');
            return;
        }

        if (isPreRecord && !recordDate) {
            alert('Пожалуйста, выберите дату и время для предварительной записи');
            return;
        }

        try {
            const requestData = {
                title: carNumber,
                comment: comment
            };

            if (isPreRecord && recordDate) {
                requestData.record = new Date(recordDate).toISOString();
            }

            console.log('Sending request:', requestData);
            const response = await axios.post('/api/AddRecord', requestData);
            console.log('Response received:', response.data);
            
            if (response.data.message === 'Record added successfully') {
                this.showSuccessModal(carNumber, isPreRecord, recordDate);
                this.clearForm();
                this.loadQueue();
            } else {
                throw new Error(response.data.error || 'Unknown error');
            }
        } catch (error) {
            console.error('Ошибка при получении талона:', error);
            
            let errorMessage = 'Ошибка при получении талона. Попробуйте еще раз.';
            if (error.response && error.response.data && error.response.data.error) {
                errorMessage = error.response.data.error;
            }
            
            alert(errorMessage);
        }
    }

    showSuccessModal(carNumber, isPreRecord, recordDate) {
        this.ticketNumber.textContent = `Автомобиль: ${carNumber}`;
        
        if (isPreRecord && recordDate) {
            const date = new Date(recordDate);
            this.ticketInfo.textContent = `Запись на: ${date.toLocaleString('ru-RU')}`;
        } else {
            this.ticketInfo.textContent = 'Текущая очередь - ожидайте вызова';
        }
        
        this.ticketModal.style.display = 'flex';
    }

    closeModal() {
        this.ticketModal.style.display = 'none';
    }

    clearForm() {
        this.carNumberInput.value = '';
        this.commentInput.value = '';
        this.preRecordCheckbox.checked = false;
        this.preRecordFields.style.display = 'none';
        this.recordDateInput.value = '';
    }

    showError(message) {
        this.inWorkList.innerHTML = `<div class="error-message">${message}</div>`;
        this.queueList.innerHTML = `<div class="error-message">${message}</div>`;
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    new TireService();
});