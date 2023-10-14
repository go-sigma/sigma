import React from 'react';

const DevelopBg = ({ color = '#161F31', ...rest }: any) => (
  <svg width="130" height="48" xmlns="http://www.w3.org/2000/svg" {...rest}>
    <g
      stroke={color}
      strokeWidth="1.04"
      fill="none"
      fillRule="evenodd"
      opacity=".4"
      strokeLinecap="round"
      strokeLinejoin="round">
      <path
        d="M.79 1.308v20.178h13.865M51.773 39.254v-8.89h78.193"
        strokeDasharray="4.530140972137452,4.530140972137452"
      />
      <path d="M66.356 40.215v1.474-1.474zM62.811 43.76h1.474-1.474zM66.356 47.305v-1.474 1.474zM69.901 43.76h-1.474 1.474zM52.695 7.247v.982-.982zM50.334 9.608h.982-.982zM52.695 11.968v-.981.981zM55.056 9.608h-.982.982z" />
    </g>
  </svg>
);

export default DevelopBg;
