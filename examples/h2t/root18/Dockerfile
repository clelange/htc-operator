FROM continuumio/miniconda3:4.7.12

ARG ROOT_VERSION=6.18.04
ARG PYTHON_VERSION=3.7

RUN conda config --add channels conda-forge && \
    conda config --set allow_softlinks false && \
    conda config --set always_copy true && \
    conda install --yes --quiet python=$PYTHON_VERSION root=$ROOT_VERSION && \
    conda clean --yes --all --force-pkgs-dirs
