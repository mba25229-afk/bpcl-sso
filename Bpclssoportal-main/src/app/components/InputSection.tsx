import { Search, Calendar, Upload, ChevronLeft, ChevronRight, History, RefreshCw, CheckCircle, AlertCircle } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { format, startOfMonth } from 'date-fns';
import { FileHistoryModal } from './FileHistoryModal';
import { api } from '../../api/client';

interface FileHistoryItem {
  id: string;
  name: string;
  uploadDate: Date;
  size: string;
}

export function InputSection({ onFetch }: { onFetch: (ccNumber: string, selectedDate: Date, file?: File) => void }) {
  const [ccNumber, setCcNumber] = useState('');
  const [selectedDate, setSelectedDate] = useState(startOfMonth(new Date()));
  const [showCalendar, setShowCalendar] = useState(false);
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const [fileHistory, setFileHistory] = useState<FileHistoryItem[]>([]);
  const [showHistory, setShowHistory] = useState(false);
  const [etlLoading, setEtlLoading] = useState(false);
  const [etlStatus, setEtlStatus] = useState<'idle' | 'success' | 'error'>('idle');
  const [etlError, setEtlError] = useState<string | null>(null);

  const calendarRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const currentYear = selectedDate.getFullYear();
  const years = Array.from({ length: 5 }, (_, i) => currentYear - 2 + i);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (calendarRef.current && !calendarRef.current.contains(event.target as Node)) {
        setShowCalendar(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleFetch = () => {
    if (ccNumber) {
      const cleanCC = ccNumber.replace(/^CC\s*/i, '').trim();
      onFetch(cleanCC, selectedDate, uploadedFile || undefined);
    }
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setUploadedFile(file);
      setFileHistory([
        { id: Date.now().toString(), name: file.name, uploadDate: new Date(), size: formatFileSize(file.size) },
        ...fileHistory,
      ]);
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  const handleDeleteHistory = (id: string) => {
    setFileHistory(fileHistory.filter(item => item.id !== id));
  };

  const handleETLSync = async () => {
    setEtlLoading(true);
    setEtlStatus('idle');
    setEtlError(null);
    try {
      await api.triggerETL();
      setEtlStatus('success');
      setTimeout(() => setEtlStatus('idle'), 3000);
    } catch (err: any) {
      setEtlStatus('error');
      setEtlError(err.message || 'ETL sync failed');
    } finally {
      setEtlLoading(false);
    }
  };

  const selectMonth = (month: number) => {
    setSelectedDate(new Date(selectedDate.getFullYear(), month, 1));
  };

  const selectYear = (year: number) => {
    setSelectedDate(new Date(year, selectedDate.getMonth(), 1));
  };

  const handleMonthChange = (increment: number) => {
    const d = new Date(selectedDate);
    d.setMonth(d.getMonth() + increment);
    setSelectedDate(startOfMonth(d));
  };

  return (
    <div className="max-w-5xl mx-auto">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-8">
        <div className="flex items-end gap-4">
          {/* CC Number */}
          <div className="flex-1">
            <label className="block text-sm text-gray-700 mb-2">CC Number</label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
              <input
                type="text"
                placeholder="Enter CC Number"
                value={ccNumber}
                onChange={(e) => setCcNumber(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleFetch()}
                className="w-full pl-11 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none"
                onFocus={(e) => (e.target.style.boxShadow = '0 0 0 2px rgba(0,123,201,0.2)')}
                onBlur={(e) => (e.target.style.boxShadow = '')}
              />
            </div>
          </div>

          {/* Month + Year picker */}
          <div className="flex-1 relative" ref={calendarRef}>
            <label className="block text-sm text-gray-700 mb-2">Month</label>
            <div className="relative">
              <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
              <button
                type="button"
                onClick={() => setShowCalendar(!showCalendar)}
                className="w-full pl-11 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none bg-white cursor-pointer text-left"
                onFocus={(e) => (e.target.style.boxShadow = '0 0 0 2px rgba(0,123,201,0.2)')}
                onBlur={(e) => (e.target.style.boxShadow = '')}
              >
                {format(selectedDate, 'MMMM yyyy')}
              </button>
            </div>

            {showCalendar && (
              <div className="absolute top-full mt-2 bg-white rounded-lg shadow-lg border border-gray-200 p-4 z-10 w-72">
                {/* Year navigation */}
                <div className="flex items-center justify-between mb-3">
                  <button onClick={() => handleMonthChange(-12)} className="p-1 hover:bg-gray-100 rounded">
                    <ChevronLeft size={18} className="text-gray-600" />
                  </button>
                  <select
                    value={currentYear}
                    onChange={(e) => selectYear(parseInt(e.target.value))}
                    className="px-3 py-1 rounded-lg text-white text-sm"
                    style={{ backgroundColor: '#007BC9', border: 'none' }}
                  >
                    {years.map(y => <option key={y} value={y}>{y}</option>)}
                  </select>
                  <button onClick={() => handleMonthChange(12)} className="p-1 hover:bg-gray-100 rounded">
                    <ChevronRight size={18} className="text-gray-600" />
                  </button>
                </div>

                {/* Month grid */}
                <div className="grid grid-cols-3 gap-1">
                  {['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'].map((m, i) => {
                    const isSelected = selectedDate.getMonth() === i && selectedDate.getFullYear() === currentYear;
                    return (
                      <button
                        key={m}
                        onClick={() => { selectMonth(i); setShowCalendar(false); }}
                        className="py-2 text-sm rounded-lg transition-colors"
                        style={isSelected ? { backgroundColor: '#007BC9', color: 'white' } : {}}
                        onMouseEnter={(e) => { if (!isSelected) (e.currentTarget.style.backgroundColor = '#f3f4f6'); }}
                        onMouseLeave={(e) => { if (!isSelected) (e.currentTarget.style.backgroundColor = ''); }}
                      >
                        {m}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
          </div>

          {/* Fetch button */}
          <button
            onClick={handleFetch}
            className="px-8 py-3 rounded-lg text-gray-900 transition-all hover:shadow-md disabled:opacity-50 disabled:cursor-not-allowed"
            style={{ backgroundColor: '#FFE000' }}
            disabled={!ccNumber}
          >
            Fetch Data
          </button>

          {/* Sync from Google Sheets */}
          <div className="flex items-center gap-2">
            {etlStatus === 'success' ? (
              <div className="flex items-center gap-2 px-4 py-3 bg-green-50 border border-green-200 rounded-lg">
                <CheckCircle size={20} className="text-green-600" />
                <span className="text-green-700 text-sm">Synced!</span>
              </div>
            ) : etlStatus === 'error' ? (
              <div className="flex items-center gap-2 px-4 py-3 bg-red-50 border border-red-200 rounded-lg">
                <AlertCircle size={20} className="text-red-600" />
                <span className="text-red-700 text-sm truncate max-w-[150px]">{etlError || 'Failed'}</span>
              </div>
            ) : null}
            <button
              onClick={handleETLSync}
              disabled={etlLoading}
              className="px-4 py-3 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors flex items-center gap-2 disabled:opacity-50"
            >
              {etlLoading ? (
                <RefreshCw size={20} className="animate-spin" />
              ) : (
                <RefreshCw size={20} />
              )}
              <span>{etlLoading ? 'Syncing...' : 'Sync from Google'}</span>
            </button>
          </div>
        </div>
      </div>

      {showHistory && (
        <FileHistoryModal
          onClose={() => setShowHistory(false)}
          history={fileHistory}
          onDelete={handleDeleteHistory}
        />
      )}
    </div>
  );
}
