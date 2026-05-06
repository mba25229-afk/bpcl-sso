import { Search, Calendar, Upload, ChevronLeft, ChevronRight, History, RefreshCw, CheckCircle } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { format, startOfMonth, endOfMonth, startOfWeek, endOfWeek, addDays, isSameMonth, isSameDay } from 'date-fns';
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
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [showCalendar, setShowCalendar] = useState(false);
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const [fileHistory, setFileHistory] = useState<FileHistoryItem[]>([]);
  const [showHistory, setShowHistory] = useState(false);

  // Delhi Master upload state
  const [uploadType, setUploadType] = useState<'performance' | 'delhi_master'>('performance');
  const [isUploading, setIsUploading] = useState(false);
  const [uploadSuccess, setUploadSuccess] = useState(false);
  const [recomputeLoading, setRecomputeLoading] = useState(false);
  const [recomputeResult, setRecomputeResult] = useState<any>(null);

  const calendarRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

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

  const handleMonthChange = (increment: number) => {
    const newDate = new Date(currentMonth);
    newDate.setMonth(newDate.getMonth() + increment);
    setCurrentMonth(newDate);
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setUploadedFile(file);

      const newHistoryItem: FileHistoryItem = {
        id: Date.now().toString(),
        name: file.name,
        uploadDate: new Date(),
        size: formatFileSize(file.size),
      };
      setFileHistory([newHistoryItem, ...fileHistory]);
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

  const COMPETITION_ID = 'c0000000-0001-0001-0001-000000000001';

  const handleDelhiMasterUpload = async () => {
    if (!uploadedFile) return;
    setIsUploading(true);
    setUploadSuccess(false);
    setRecomputeResult(null);

    try {
      const formData = new FormData();
      formData.append('file', uploadedFile);
      await api.uploadDelhiMaster(formData);
      setUploadSuccess(true);
    } catch (err) {
      console.error('Delhi Master upload failed:', err);
      alert('Failed to upload Delhi Master file. Please try again.');
    } finally {
      setIsUploading(false);
    }
  };

  const handleRecompute = async () => {
    setRecomputeLoading(true);
    setRecomputeResult(null);

    try {
      const result = await api.recomputeScores(COMPETITION_ID);
      setRecomputeResult(result);
    } catch (err) {
      console.error('Recompute failed:', err);
      alert('Failed to recompute scores. Please try again.');
    } finally {
      setRecomputeLoading(false);
    }
  };

  const getDaysInMonth = () => {
    const monthStart = startOfMonth(currentMonth);
    const monthEnd = endOfMonth(currentMonth);
    const startDate = startOfWeek(monthStart);
    const endDate = endOfWeek(monthEnd);

    const days = [];
    let day = startDate;

    while (day <= endDate) {
      days.push(day);
      day = addDays(day, 1);
    }

    return days;
  };

  const days = getDaysInMonth();
  const currentYear = currentMonth.getFullYear();
  const years = Array.from({ length: 5 }, (_, i) => currentYear - 2 + i);

  return (
    <div className="max-w-5xl mx-auto">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-8">
        {/* Upload Type Toggle */}
        <div className="mb-6 flex items-center gap-4">
          <span className="text-sm text-gray-700 font-medium">Upload Type:</span>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => {
                setUploadType('performance');
                setUploadedFile(null);
                setUploadSuccess(false);
                setRecomputeResult(null);
              }}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                uploadType === 'performance'
                  ? 'text-white'
                  : 'text-gray-700 bg-gray-100 hover:bg-gray-200'
              }`}
              style={uploadType === 'performance' ? { backgroundColor: '#007BC9' } : {}}
            >
              ● Performance Data
            </button>
            <button
              type="button"
              onClick={() => {
                setUploadType('delhi_master');
                setUploadedFile(null);
                setUploadSuccess(false);
                setRecomputeResult(null);
              }}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                uploadType === 'delhi_master'
                  ? 'text-white'
                  : 'text-gray-700 bg-gray-100 hover:bg-gray-200'
              }`}
              style={uploadType === 'delhi_master' ? { backgroundColor: '#007BC9' } : {}}
            >
              ○ Delhi Master (Market Share)
            </button>
          </div>
        </div>

        {/* Delhi Master Helper Text */}
        {uploadType === 'delhi_master' && (
          <div className="mb-4 p-3 rounded-lg" style={{ backgroundColor: '#eff6ff' }}>
            <p className="text-sm text-gray-700">
              Upload the Delhi Master file containing all OMC volumes by trading area. Supported format: .xlsx
            </p>
          </div>
        )}

        {/* Conditional Fields based on Upload Type */}
        <div className="flex items-end gap-4">
          {uploadType === 'performance' ? (
            <>
              <div className="flex-1">
                <label className="block text-sm text-gray-700 mb-2">CC Number</label>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
                  <input
                    type="text"
                    placeholder="Enter CC Number"
                    value={ccNumber}
                    onChange={(e) => setCcNumber(e.target.value)}
                    className="w-full pl-11 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:border-transparent"
                    onFocus={(e) => e.target.style.boxShadow = '0 0 0 2px rgba(0, 123, 201, 0.2)'}
                    onBlur={(e) => e.target.style.boxShadow = ''}
                  />
                </div>
              </div>

              <div className="flex-1 relative" ref={calendarRef}>
                <label className="block text-sm text-gray-700 mb-2">Date</label>
                <div className="relative">
                  <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
                  <button
                    type="button"
                    onClick={() => setShowCalendar(!showCalendar)}
                    className="w-full pl-11 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 appearance-none bg-white cursor-pointer text-left"
                    onFocus={(e) => e.target.style.boxShadow = '0 0 0 2px rgba(0, 123, 201, 0.2)'}
                    onBlur={(e) => e.target.style.boxShadow = ''}
                  >
                    {format(selectedDate, 'dd MMMM yyyy')}
                  </button>
                </div>

                {showCalendar && (
                  <div className="absolute top-full mt-2 bg-white rounded-lg shadow-lg border border-gray-200 p-4 z-10 w-96">
                    <div className="flex items-center justify-between mb-4">
                      <button
                        onClick={() => handleMonthChange(-1)}
                        className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                      >
                        <ChevronLeft size={20} className="text-gray-600" />
                      </button>
                      <div className="flex items-center gap-2">
                        <select
                          value={currentMonth.getMonth()}
                          onChange={(e) => {
                            const newDate = new Date(currentMonth);
                            newDate.setMonth(parseInt(e.target.value));
                            setCurrentMonth(newDate);
                          }}
                          className="px-3 py-1 rounded-lg text-white"
                          style={{ backgroundColor: '#007BC9', border: 'none' }}
                        >
                          {['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'].map((month, index) => (
                            <option key={month} value={index}>{month}</option>
                          ))}
                        </select>
                        <select
                          value={currentYear}
                          onChange={(e) => {
                            const newDate = new Date(currentMonth);
                            newDate.setFullYear(parseInt(e.target.value));
                            setCurrentMonth(newDate);
                          }}
                          className="px-3 py-1 rounded-lg text-white"
                          style={{ backgroundColor: '#007BC9', border: 'none' }}
                        >
                          {years.map(year => (
                            <option key={year} value={year}>{year}</option>
                          ))}
                        </select>
                      </div>
                      <button
                        onClick={() => handleMonthChange(1)}
                        className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                      >
                        <ChevronRight size={20} className="text-gray-600" />
                      </button>
                    </div>

                    <div className="grid grid-cols-7 gap-1 mb-2">
                      {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map(day => (
                        <div key={day} className="text-center text-xs text-gray-600 py-2">
                          {day}
                        </div>
                      ))}
                    </div>

                    <div className="grid grid-cols-7 gap-1">
                      {days.map((day, index) => {
                        const isSelected = isSameDay(day, selectedDate);
                        const isToday = isSameDay(day, new Date());
                        const isCurrentMonth = isSameMonth(day, currentMonth);

                        return (
                          <button
                            key={index}
                            onClick={() => {
                              setSelectedDate(day);
                              setShowCalendar(false);
                            }}
                            className={`p-2 text-sm rounded-lg transition-colors ${
                              isSelected
                                ? 'text-white'
                                : isToday
                                ? 'bg-gray-100 text-gray-900'
                                : isCurrentMonth
                                ? 'text-gray-700 hover:bg-gray-100'
                                : 'text-gray-400 hover:bg-gray-50'
                            }`}
                            style={isSelected ? { backgroundColor: '#007BC9' } : {}}
                          >
                            {format(day, 'd')}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                )}
              </div>

              <button
                onClick={handleFetch}
                className="px-8 py-3 rounded-lg text-gray-900 transition-all hover:shadow-md disabled:opacity-50 disabled:cursor-not-allowed"
                style={{ backgroundColor: '#FFE000' }}
                disabled={!ccNumber}
              >
                Fetch Data
              </button>
            </>
          ) : (
            /* Delhi Master Upload */
            <div className="flex items-center gap-2">
              <input
                ref={fileInputRef}
                type="file"
                accept=".xlsx,.xls"
                onChange={handleFileUpload}
                className="hidden"
              />
              <button
                onClick={() => fileInputRef.current?.click()}
                className="px-4 py-3 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors flex items-center gap-2"
                disabled={isUploading}
              >
                <Upload size={20} />
                <span>{uploadedFile ? uploadedFile.name.slice(0, 20) + '...' : 'Select Delhi Master'}</span>
              </button>
              {uploadedFile && !uploadSuccess && (
                <button
                  onClick={handleDelhiMasterUpload}
                  className="px-4 py-3 rounded-lg text-white transition-all hover:shadow-md disabled:opacity-50"
                  style={{ backgroundColor: '#007BC9' }}
                  disabled={isUploading}
                >
                  {isUploading ? 'Uploading...' : 'Upload'}
                </button>
              )}
              {uploadSuccess && (
                <span className="flex items-center gap-1 text-green-600 text-sm">
                  <CheckCircle size={16} />
                  Delhi Master uploaded. Processing market share data...
                </span>
              )}
            </div>
          )}

          {/* Performance Upload Section - only show when performance mode */}
          {uploadType === 'performance' && (
            <div className="flex items-center gap-2">
              <input
                ref={fileInputRef}
                type="file"
                accept=".xlsx,.xls,.csv"
                onChange={handleFileUpload}
                className="hidden"
              />
              <button
                onClick={() => fileInputRef.current?.click()}
                className="px-4 py-3 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors flex items-center gap-2"
              >
                <Upload size={20} />
                <span>{uploadedFile ? uploadedFile.name.slice(0, 15) + '...' : 'Upload Excel'}</span>
              </button>
              <button
                onClick={() => setShowHistory(true)}
                className="p-3 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors relative"
                title="File History"
              >
                <History size={20} />
                {fileHistory.length > 0 && (
                  <span className="absolute -top-1 -right-1 w-5 h-5 rounded-full text-xs flex items-center justify-center text-white" style={{ backgroundColor: '#007BC9' }}>
                    {fileHistory.length}
                  </span>
                )}
              </button>
            </div>
          )}
        </div>

        {/* Delhi Master Recompute Section */}
        {uploadType === 'delhi_master' && uploadSuccess && (
          <div className="mt-4 p-4 rounded-lg" style={{ backgroundColor: '#eff6ff' }}>
            <div className="flex items-center justify-between">
              <div>
                {recomputeResult ? (
                  <p className="text-gray-900 font-medium">
                    🏆 Scores updated. Top dealer: {recomputeResult.top10_after?.[0]?.outlet_name ?? 'N/A'} — {recomputeResult.top10_after?.[0]?.total_score?.toFixed(2) ?? '0'} pts
                  </p>
                ) : (
                  <p className="text-gray-700">
                    Delhi Master uploaded and processed successfully.
                  </p>
                )}
              </div>
              <button
                onClick={handleRecompute}
                className="px-4 py-2 rounded-lg text-white flex items-center gap-2 transition-all hover:shadow-md disabled:opacity-50"
                style={{ backgroundColor: '#007BC9' }}
                disabled={recomputeLoading}
              >
                <RefreshCw size={16} className={recomputeLoading ? 'animate-spin' : ''} />
                {recomputeLoading ? 'Recalculating...' : 'Recompute Competition Scores'}
              </button>
            </div>
          </div>
        )}
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