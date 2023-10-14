import React from 'react';

const Hamburger = ({ color = '#161F31', ...rest }: any) => (
  <svg width="24" height="18" xmlns="http://www.w3.org/2000/svg" {...rest}>
    <g
      stroke={color}
      strokeWidth="1.65"
      fill="none"
      fillRule="evenodd"
      strokeDasharray="4.4"
      strokeLinecap="round"
      strokeLinejoin="round">
      <path d="M1 16.95h24.2M1 9.25h24.2M1 1.55h24.2" />
    </g>
  </svg>
);

export default Hamburger;
