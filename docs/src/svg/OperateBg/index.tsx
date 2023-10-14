import React from 'react';

const Operate = ({ color = '#161F31', ...rest }: any) => (
  <svg width="30" height="36" xmlns="http://www.w3.org/2000/svg" {...rest}>
    <g
      stroke={color}
      strokeWidth="1.01"
      fill="none"
      fillRule="evenodd"
      strokeLinecap="round"
      strokeLinejoin="round">
      <path d="M1 7.928h12.45V1.28H1zM1 17.899h12.45V11.25H1zM1 34.518h12.45V27.87H1zM16.672 26.208h12.45V19.56h-12.45zM7.227 17.899v9.971V17.9zM7.227 7.928v3.323-3.323zM16.672 22.884H7.227h9.445zM13.45 14.575h9.45v4.985z" />
    </g>
  </svg>
);

export default Operate;
