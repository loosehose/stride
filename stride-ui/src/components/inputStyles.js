import { useState } from 'react';

export const useInputStyles = (initialState = false) => {
    const [isFocused, setIsFocused] = useState(initialState);

    const handleFocus = () => {
        setIsFocused(true);
    };

    const handleBlur = () => {
        setIsFocused(false);
    };

    const getInputStyles = () => ({
        backgroundColor: '#272a3d',
        borderColor: isFocused ? 'rgba(93, 37, 85, 0.8)' : 'rgba(82, 95, 127, 0.8)',
        borderRadius: '.25rem',
        borderWidth: '1px',
        boxShadow: isFocused ? '0 0 0 0.2rem rgba(93, 37, 85, 0.25)' : 'none',
        color: 'white',
        minHeight: 'calc(1.5em + .75rem + 4px)',
        fontSize: '0.875rem',
        fontFamily: '"Poppins", sans-serif',
        transition: 'border-color .15s ease-in-out, box-shadow .15s ease-in-out',
        '&:hover': {
            borderColor: 'rgba(82, 95, 127, 1)',
        },
    });

    return { isFocused, handleFocus, handleBlur, getInputStyles };
};