import {useState} from 'react';
import './App.css';
//import {} from "../wailsjs/go/main/App";

function App () {
    const [hoveredFeature, setHoveredFeature] = useState(null);
    const [buttonClicks, setButtonClicks] = useState({ primary: 0, secondary: 0 });
  
    const features = [
      {
        id: 1,
        icon: '🎨',
        title: 'Modern Design',
        description: 'Beautiful gradients, shadows, and smooth transitions create a professional look.'
      },
      {
        id: 2,
        icon: '📱',
        title: 'Responsive',
        description: 'Automatically adapts to different screen sizes and devices.'
      },
      {
        id: 3,
        icon: '⚡',
        title: 'Interactive',
        description: 'React state management and event handling for dynamic experiences.'
      },
      {
        id: 4,
        icon: '🎯',
        title: 'Component-Based',
        description: 'Reusable, maintainable code with modern React patterns.'
      }
    ];
  
    const handleButtonClick = (type) => {
      setButtonClicks(prev => ({
        ...prev,
        [type]: prev[type] + 1
      }));
    };
  
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-500 via-purple-600 to-purple-800">
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          {/* Main Card */}
          <div className="bg-white rounded-2xl shadow-2xl overflow-hidden mb-8 transform hover:-translate-y-2 transition-all duration-300">
            {/* Header */}
            <div className="bg-gradient-to-r from-red-400 to-pink-500 text-white px-8 py-12 text-center">
              <h1 className="text-4xl md:text-5xl font-light mb-4">
                Welcome to React
              </h1>
              <p className="text-xl opacity-90">
                A modern, interactive React component with state management
              </p>
            </div>
  
            {/* Content */}
            <div className="p-8">
              {/* About Section */}
              <div className="mb-12">
                <h2 className="text-2xl font-semibold text-gray-700 mb-6 border-l-4 border-blue-500 pl-4">
                  About This Component
                </h2>
                <p className="text-gray-600 mb-4 leading-relaxed">
                  This is a React page component that demonstrates modern React patterns including 
                  state management, event handling, and component composition. It's built with 
                  Tailwind CSS for responsive styling.
                </p>
                <p className="text-gray-600 leading-relaxed">
                  The component features interactive elements, hover effects, and dynamic state 
                  updates that showcase React's capabilities for building engaging user interfaces.
                </p>
              </div>
  
              {/* Features Grid */}
              <div className="mb-12">
                <h2 className="text-2xl font-semibold text-gray-700 mb-6 border-l-4 border-blue-500 pl-4">
                  Key Features
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {features.map((feature) => (
                    <div
                      key={feature.id}
                      className={`bg-gray-50 p-6 rounded-xl border-l-4 border-red-400 transition-all duration-300 cursor-pointer ${
                        hoveredFeature === feature.id 
                          ? 'bg-gray-100 transform translate-x-2 shadow-md' 
                          : 'hover:bg-gray-100 hover:transform hover:translate-x-1'
                      }`}
                      onMouseEnter={() => setHoveredFeature(feature.id)}
                      onMouseLeave={() => setHoveredFeature(null)}
                    >
                      <div className="flex items-center mb-3">
                        <span className="text-2xl mr-3">{feature.icon}</span>
                        <h3 className="text-xl font-semibold text-gray-800">
                          {feature.title}
                        </h3>
                      </div>
                      <p className="text-gray-600 text-sm leading-relaxed">
                        {feature.description}
                      </p>
                      {hoveredFeature === feature.id && (
                        <div className="mt-3 text-xs text-blue-600 font-medium animate-pulse">
                          Feature #{feature.id} - Click to explore more
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
  
              {/* Interactive Section */}
              <div className="mb-8">
                <h2 className="text-2xl font-semibold text-gray-700 mb-6 border-l-4 border-blue-500 pl-4">
                  Interactive Demo
                </h2>
                <p className="text-gray-600 mb-6 leading-relaxed">
                  These buttons demonstrate React state management. Click them to see the counters update!
                </p>
                
                <div className="flex flex-wrap gap-4 mb-6">
                  <button
                    onClick={() => handleButtonClick('primary')}
                    className="bg-gradient-to-r from-blue-500 to-purple-600 text-white px-8 py-3 rounded-full font-medium transition-all duration-300 hover:-translate-y-1 hover:shadow-lg hover:shadow-blue-500/25 transform active:scale-95"
                  >
                    Primary Action ({buttonClicks.primary})
                  </button>
                  <button
                    onClick={() => handleButtonClick('secondary')}
                    className="bg-gradient-to-r from-gray-500 to-gray-600 text-white px-8 py-3 rounded-full font-medium transition-all duration-300 hover:-translate-y-1 hover:shadow-lg hover:shadow-gray-500/25 transform active:scale-95"
                  >
                    Secondary Action ({buttonClicks.secondary})
                  </button>
                </div>
  
                {(buttonClicks.primary > 0 || buttonClicks.secondary > 0) && (
                  <div className="bg-green-50 border border-green-200 rounded-lg p-4 animate-fadeIn">
                    <p className="text-green-800 font-medium">
                      🎉 State Updated! Primary: {buttonClicks.primary}, Secondary: {buttonClicks.secondary}
                    </p>
                    <button
                      onClick={() => setButtonClicks({ primary: 0, secondary: 0 })}
                      className="text-green-600 text-sm underline mt-2 hover:text-green-800"
                    >
                      Reset counters
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
  
          {/* Footer */}
          <div className="text-center text-white bg-black bg-opacity-10 rounded-xl p-6">
            <p className="opacity-90">
              &copy; 2025 React Page Component. Built with React & Tailwind CSS
            </p>
            <p className="text-sm opacity-75 mt-2">
              Total interactions: {buttonClicks.primary + buttonClicks.secondary}
            </p>
          </div>
        </div>
  
        <style jsx>{`
          @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
          }
          .animate-fadeIn {
            animation: fadeIn 0.3s ease-out;
          }
        `}</style>
      </div>
    );
  };
  

export default App
