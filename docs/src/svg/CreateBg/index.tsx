import React from 'react';

const CreateBg = ({ color = '#161F31', ...rest }: any) => (
  <svg width="56" height="63" xmlns="http://www.w3.org/2000/svg" {...rest}>
    <g
      stroke={color}
      strokeWidth="1.045"
      fill="none"
      fillRule="evenodd"
      opacity=".4"
      strokeLinecap="round"
      strokeLinejoin="round">
      <g strokeDasharray="4.554624032974243,4.554624032974243">
        <path d="M27.195.605L27.15 19.25M45.647 53.553h9.68v-19.36" />
      </g>
      <path d="M5.686 6.832V8.78M1 11.517h1.947M5.686 16.201v-1.948M10.368 11.517H8.42" />
      <g>
        <path d="M32 58v1M30 60h1M32 62v-1M34 60h-1" />
      </g>
    </g>
  </svg>
);

export default CreateBg;
