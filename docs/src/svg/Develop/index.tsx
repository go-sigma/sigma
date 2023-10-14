import React from 'react';

const Develop = ({ color = '#161F31', ...rest }: any) => (
  <svg width="23" height="31" xmlns="http://www.w3.org/2000/svg" {...rest}>
    <g fill="none" fillRule="evenodd">
      <path
        d="M22.205 5.931h-5.072V.86l5.072 5.071zm0 23.942H.635V.86h16.498l5.072 5.071v23.942z"
        stroke={color}
        strokeWidth="1.06"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        fill={color}
        d="M4.717 16.514v.06l3.139 1.57v1.45l-4.528-2.295v-1.51l4.528-2.294v1.45l-3.139 1.57M14.136 10.537l-3.803 9.66h-1.63l3.804-9.66h1.629M18.123 16.575v-.06l-3.142-1.57v-1.45l4.53 2.295v1.509l-4.53 2.294v-1.449l3.142-1.57"
      />
    </g>
  </svg>
);

export default Develop;
