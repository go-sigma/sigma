import clsx from 'clsx';
import React from 'react';

import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Typed from '@site/src/components/Typed';
import SvgHero from '@site/src/svg/Hero';
import SvgCreate from '@site/src/svg/Create';
import SvgCreateBg from '@site/src/svg/CreateBg';
import SvgDevelop from '@site/src/svg/Develop';
import SvgDevelopBg from '@site/src/svg/DevelopBg';
import SvgExplore from '@site/src/svg/Explore';
import SvgOperate from '@site/src/svg/Operate';
import SvgExploreBg from '@site/src/svg/ExploreBg';

function Hero() {
  return (
    <header className="rds-hero">
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <div className="row">
              <div className="col col--8">
                <h1 className="hero-title">
                  The home of
                  <br /> sigma developers
                </h1>

                <h2 className="hero-subtitle">
                  <Typed
                    strings={['>_ Made by developers for you by heart']}
                    typeSpeed={75}
                  />
                </h2>
                <Link
                  className={clsx(
                    'button button--outline button--secondary button--lg hero-started',
                  )}
                  to={useBaseUrl('docs/sigma')}>
                  Get started
                </Link>
                <Link
                  className={clsx(
                    'button button--outline button--secondary button--lg hero-try-demo',
                  )}
                  to={"https://sigma.tosone.cn"}>
                  Try demo
                </Link>
              </div>
              <div className="col col--4">
                <SvgHero color="#FFFFFF" className="illustration" />
              </div>
            </div>
            <div className="boxes">
              <div className="box box-create">
                <SvgCreateBg color="#FFFFFF" className="bg" />
                <span className="icon">
                  <SvgCreate color="#FFFFFF" />
                </span>
                <div className="text">
                  <h3 className="title">Store</h3>
                  <p className="description">
                    Support store OCI artifact(include oci v1 and docker schema v2), multiarch image, helm chart, cnab, apptainer...
                  </p>
                </div>
              </div>

              <div className="box box-develop">
                <SvgDevelopBg color="#FFFFFF" className="bg" />
                <span className="icon">
                  <SvgDevelop color="#FFFFFF" />
                </span>
                <div className="text">
                  <h3 className="title">Develop</h3>
                  <p className="description">
                    All of api and authorization service in one service.
                  </p>
                </div>
              </div>

              <div className="box box-explore">
                <SvgExploreBg color="#FFFFFF" className="bg" />
                <span className="icon">
                  <SvgExplore color="#FFFFFF" />
                </span>
                <div className="text">
                  <h3 className="title">P2P</h3>
                  <p className="description">
                    P2P distribute blobs with dragonfly.
                  </p>
                </div>
              </div>

              <div className="box box-operate">
                <SvgExploreBg color="#FFFFFF" className="bg" />
                <span className="icon">
                  <SvgOperate />
                </span>
                <div className="text">
                  <h3 className="title">Distribute</h3>
                  <p className="description">
                    Distribute artifact to another instance.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

export default Hero;
