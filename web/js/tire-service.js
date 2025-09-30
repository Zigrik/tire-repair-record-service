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

        this.availableSlotsCache = {};
        this.workStations = 3; // Количество рабочих постов
        this.averageServiceTime = 40; // Среднее время обслуживания в минутах

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
        this.recordDateInput.addEventListener('focus', () => this.loadAvailableSlots());

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

    async togglePreRecord() {
        if (this.preRecordCheckbox.checked) {
            this.preRecordFields.style.display = 'block';
            await this.loadAvailableSlots();
        } else {
            this.preRecordFields.style.display = 'none';
            this.recordDateInput.value = '';
        }
    }

    async loadAvailableSlots() {
        try {
            const today = new Date().toISOString().split('T')[0];
            const response = await axios.post('/api/GetAvailableSlots', { date: today });
            const slots = response.data.slots || [];
            
            // Получаем текущую очередь для расчета времени
            const queueResponse = await axios.get('/api/GetTodayRecords');
            const waitingRecords = (queueResponse.data.records || []).filter(r => 
                r.status === 'wait' && !r.record
            );
            
            const availableSlots = this.calculateRealAvailableSlots(slots, waitingRecords.length);
            this.updateDateTimeInput(availableSlots);
            
        } catch (error) {
            console.error('Ошибка загрузки доступных слотов:', error);
        }
    }

    calculateRealAvailableSlots(slots, waitingCount) {
        const now = new Date();
        const minStartTime = new Date(now.getTime() + this.calculateMinStartTime(waitingCount) * 60000);
        
        return slots.filter(slot => {
            const slotTime = new Date(slot);
            return slotTime >= minStartTime;
        });
    }

    calculateMinStartTime(waitingCount) {
        // Время начала = (количество в очереди * среднее время) / количество постов
        return Math.ceil((waitingCount * this.averageServiceTime) / this.workStations);
    }

    updateDateTimeInput(availableSlots) {
        // Создаем datalist с доступными слотами
        let datalist = document.getElementById('availableSlotsList');
        if (!datalist) {
            datalist = document.createElement('datalist');
            datalist.id = 'availableSlotsList';
            document.body.appendChild(datalist);
        }
        
        datalist.innerHTML = availableSlots.map(slot => {
            const date = new Date(slot);
            const value = date.toISOString().slice(0, 16);
            const display = date.toLocaleString('ru-RU', {
                day: '2-digit',
                month: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            });
            return `<option value="${value}">${display}</option>`;
        }).join('');
        
        this.recordDateInput.setAttribute('list', 'availableSlotsList');
        this.recordDateInput.min = new Date().toISOString().slice(0, 16);
    }

    async loadQueue() {
        try {
            const response = await axios.get('/api/GetTodayRecords');
            const records = response.data.records || [];
            this.displayQueue(records);
        } catch (error) {
            console.error('Ошибка загрузки очереди:', error);
            this.showError('Ошибка загрузки данных');
        }
    }

    displayQueue(records) {
        // Фильтруем записи по статусам
        const inWorkRecords = records.filter(record => 
            record.status === 'in work' || record.status === 'welcome'
        );
        
        const waitingRecords = records.filter(record => 
            record.status === 'wait'
        );

        // Отображаем "В работе"
        if (inWorkRecords.length > 0) {
            this.inWorkList.innerHTML = inWorkRecords.map(record => `
                <div class="queue-item in-work">
                    <div class="ticket-number">${record.ticketNumber}</div>
                    <div class="car-number">${this.escapeHtml(record.title)}</div>
                    <div class="status">${this.getStatusText(record.status)}</div>
                </div>
            `).join('');
        } else {
            this.inWorkList.innerHTML = '<div class="empty-message">Нет машин в работе</div>';
        }

        // Отображаем "Очередь"
        if (waitingRecords.length > 0) {
            this.queueList.innerHTML = waitingRecords.map((record, index) => `
                <div class="queue-item waiting">
                    <div class="ticket-number">${record.ticketNumber}</div>
                    <div class="car-number">${this.escapeHtml(record.title)}</div>
                    <div class="position">#${index + 1}</div>
                </div>
            `).join('');
        } else {
            this.queueList.innerHTML = '<div class="empty-message">Очередь пуста</div>';
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

            const response = await axios.post('/api/AddRecord', requestData);
            
            if (response.data.message === 'Record added successfully') {
                // Получаем ID новой записи для отображения номера талона
                const queueResponse = await axios.get('/api/GetTodayRecords');
                const newRecord = queueResponse.data.records.find(r => 
                    r.title === carNumber && r.status === 'wait'
                );
                
                this.showSuccessModal(newRecord?.ticketNumber || carNumber, isPreRecord, recordDate);
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

    showSuccessModal(ticketNumber, isPreRecord, recordDate) {
        this.ticketNumber.textContent = `Талон: ${ticketNumber}`;
        
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

document.addEventListener('DOMContentLoaded', () => {
    new TireService();
});