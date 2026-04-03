import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Zero Host Pollution',
    Svg: require('@site/static/img/undraw_docusaurus_mountain.svg').default,
    description: (
      <>
        Say goodbye to global installations of <code>node</code>, <code>pyenv</code>, or complex versions.
        Everything runs locked inside an ephemeral Nix sandbox.
      </>
    ),
  },
  {
    title: 'Declarative Sandboxes',
    Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
    description: (
      <>
        Define exactly what your microservices require on a single <code>derrick.yaml</code> contract.
        From Docker images to Nix packages.
      </>
    ),
  },
  {
    title: 'Fail-Fast Validation',
    Svg: require('@site/static/img/undraw_docusaurus_react.svg').default,
    description: (
      <>
        Catch missing secrets, API ports explicitly trapped in use, and missing dependencies in milliseconds before your code even starts.
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
