import { useState } from 'react';
import { Star, Send, MessageSquare } from 'lucide-react';
import { motion } from 'motion/react';

export function Feedback() {
  const [rating, setRating] = useState(0);
  const [hoveredRating, setHoveredRating] = useState(0);
  const [category, setCategory] = useState('');
  const [message, setMessage] = useState('');
  const [submitted, setSubmitted] = useState(false);

  const categories = [
    'Fuel Quality',
    'Staff Behavior',
    'Cleanliness',
    'Pricing',
    'Speed of Service',
    'Other'
  ];

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
    setTimeout(() => {
      setRating(0);
      setCategory('');
      setMessage('');
      setSubmitted(false);
    }, 3000);
  };

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Customer Feedback</h1>
        <p className="text-gray-600">Help us improve your experience at our retail outlet</p>
      </div>

      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2">
          <div className="bg-white rounded-xl p-8 shadow-sm border border-gray-100">
            <div className="flex items-center gap-3 mb-6">
              <MessageSquare size={24} className="text-[#007BC9]" />
              <h2 className="text-2xl font-bold text-gray-900">Submit Feedback</h2>
            </div>

            {submitted ? (
              <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                className="text-center py-12"
              >
                <div className="w-20 h-20 bg-green-100 rounded-full mx-auto mb-4 flex items-center justify-center">
                  <svg className="w-10 h-10 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <h3 className="text-2xl font-bold text-gray-900 mb-2">Thank You!</h3>
                <p className="text-gray-600">Your feedback has been submitted successfully.</p>
              </motion.div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-3">
                    How would you rate your overall experience?
                  </label>
                  <div className="flex gap-2">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <button
                        key={star}
                        type="button"
                        onClick={() => setRating(star)}
                        onMouseEnter={() => setHoveredRating(star)}
                        onMouseLeave={() => setHoveredRating(0)}
                        className="transition-transform hover:scale-110"
                      >
                        <Star
                          size={40}
                          className={
                            star <= (hoveredRating || rating)
                              ? 'fill-[#FFE000] text-[#FFE000]'
                              : 'text-gray-300'
                          }
                        />
                      </button>
                    ))}
                  </div>
                  {rating > 0 && (
                    <p className="mt-2 text-sm text-gray-600">
                      {rating === 1 && 'Very Poor'}
                      {rating === 2 && 'Poor'}
                      {rating === 3 && 'Average'}
                      {rating === 4 && 'Good'}
                      {rating === 5 && 'Excellent'}
                    </p>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-3">
                    What would you like to share feedback about?
                  </label>
                  <div className="grid grid-cols-3 gap-3">
                    {categories.map((cat) => (
                      <button
                        key={cat}
                        type="button"
                        onClick={() => setCategory(cat)}
                        className={`p-3 rounded-lg border-2 transition-all ${
                          category === cat
                            ? 'border-[#007BC9] bg-blue-50 text-[#007BC9]'
                            : 'border-gray-200 hover:border-gray-300'
                        }`}
                      >
                        {cat}
                      </button>
                    ))}
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-3">
                    Your Feedback
                  </label>
                  <textarea
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    placeholder="Please share your experience with us..."
                    className="w-full h-32 p-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#007BC9] resize-none"
                    required
                  />
                </div>

                <div className="flex items-center gap-4">
                  <button
                    type="submit"
                    disabled={!rating || !category || !message}
                    className="flex items-center gap-2 px-6 py-3 bg-[#007BC9] text-white rounded-lg hover:bg-[#0056A3] transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
                  >
                    <Send size={18} />
                    Submit Feedback
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setRating(0);
                      setCategory('');
                      setMessage('');
                    }}
                    className="px-6 py-3 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                  >
                    Clear
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-gradient-to-br from-[#007BC9] to-[#0056A3] rounded-xl p-6 text-white">
            <h3 className="text-xl font-bold mb-4">Overall Rating</h3>
            <div className="text-center">
              <div className="text-6xl font-bold mb-2">4.7</div>
              <div className="flex justify-center gap-1 mb-3">
                {[1, 2, 3, 4, 5].map((star) => (
                  <Star
                    key={star}
                    size={20}
                    className={star <= 4 ? 'fill-[#FFE000] text-[#FFE000]' : 'text-white/50'}
                  />
                ))}
              </div>
              <p className="text-blue-100">Based on 1,247 reviews</p>
            </div>
          </div>

          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Rating Breakdown</h3>
            <div className="space-y-3">
              {[5, 4, 3, 2, 1].map((stars) => {
                const percentage = stars === 5 ? 68 : stars === 4 ? 22 : stars === 3 ? 7 : stars === 2 ? 2 : 1;
                return (
                  <div key={stars} className="flex items-center gap-3">
                    <div className="flex gap-0.5">
                      {[...Array(stars)].map((_, i) => (
                        <Star key={i} size={12} className="fill-[#FFE000] text-[#FFE000]" />
                      ))}
                    </div>
                    <div className="flex-1 bg-gray-100 rounded-full h-2">
                      <div
                        className="bg-[#007BC9] h-2 rounded-full"
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                    <span className="text-sm text-gray-600 w-12 text-right">{percentage}%</span>
                  </div>
                );
              })}
            </div>
          </div>

          <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Recent Feedback</h3>
            <div className="space-y-4">
              <div className="border-b border-gray-100 pb-4">
                <div className="flex items-center gap-2 mb-2">
                  {[...Array(5)].map((_, i) => (
                    <Star key={i} size={14} className="fill-[#FFE000] text-[#FFE000]" />
                  ))}
                </div>
                <p className="text-sm text-gray-700 mb-1">Excellent service and clean facility!</p>
                <p className="text-xs text-gray-500">2 hours ago</p>
              </div>
              <div className="border-b border-gray-100 pb-4">
                <div className="flex items-center gap-2 mb-2">
                  {[...Array(4)].map((_, i) => (
                    <Star key={i} size={14} className="fill-[#FFE000] text-[#FFE000]" />
                  ))}
                </div>
                <p className="text-sm text-gray-700 mb-1">Staff is very helpful and polite</p>
                <p className="text-xs text-gray-500">5 hours ago</p>
              </div>
              <div>
                <div className="flex items-center gap-2 mb-2">
                  {[...Array(5)].map((_, i) => (
                    <Star key={i} size={14} className="fill-[#FFE000] text-[#FFE000]" />
                  ))}
                </div>
                <p className="text-sm text-gray-700 mb-1">Quick service, great BeCafe coffee!</p>
                <p className="text-xs text-gray-500">1 day ago</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
