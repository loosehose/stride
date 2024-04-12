export const customStyles = {
    control: (provided, state) => ({
      ...provided,
      backgroundColor: '#272a3d',
      borderColor: state.isFocused ? 'rgba(93, 37, 85, 0.8)' : 'rgba(82, 95, 127, 0.8)', // Slightly transparent
      '&:hover': {
        borderColor: state.isFocused ? 'rgba(93, 37, 85, 1)' : 'rgba(82, 95, 127, 1)', // Full color on hover
      },
      boxShadow: state.isFocused ? `0 0 0 0.2rem rgba(93, 37, 85, 0.25)` : 'none', // Using deep purple with transparency
      borderRadius: '.25rem',
      color: 'white',
      minHeight: 'calc(1.5em + .75rem + 2px)',
      fontSize: '0.875rem',
      fontFamily: '"Poppins", sans-serif',
      transition: 'border-color .15s ease-in-out, box-shadow .15s ease-in-out',
    }),
    singleValue: (provided) => ({
      ...provided,
      color: '#dee2e6',
    }),
    menu: (provided) => ({
      ...provided,
      backgroundColor: '#272a3d',
      borderColor: '#525f7f',
      borderRadius: '.25rem',
    }),
    menuList: (provided) => ({
      ...provided,
      backgroundColor: '#272a3d',
      borderColor: '#525f7f',
    }),
    option: (provided, state) => ({
      ...provided,
      backgroundColor: state.isSelected ? '#5D2555' : '#272a3d',
      color: state.isSelected ? 'white' : '#dee2e6',
      '&:hover': {
        backgroundColor: '#5D2555',
        color: 'white',
      },
    }),
    multiValue: (provided) => ({
      ...provided,
      backgroundColor: '#525f7f',
    }),
    multiValueLabel: (provided) => ({
      ...provided,
      color: 'white',
    }),
    multiValueRemove: (provided, state) => ({
      ...provided,
      color: state.isFocused ? 'white' : '#5D2555',
      ':hover': {
        backgroundColor: '#5D2555',
        color: 'white',
      },
    }),
  };